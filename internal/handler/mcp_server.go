package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/service"
	"github.com/atbeta/picfast/internal/sqlc"
)

// MCPServerFactory creates and configures an MCP server for picfast.
type MCPServerFactory struct {
	DB     *sqlc.Queries
	Pool   *pgxpool.Pool
	Config *config.Config
}

func NewMCPServerFactory(db *sqlc.Queries, pool *pgxpool.Pool, cfg *config.Config) *MCPServerFactory {
	return &MCPServerFactory{DB: db, Pool: pool, Config: cfg}
}

func (f *MCPServerFactory) CreateServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "picfast",
		Version: "1.0.0",
	}, &mcp.ServerOptions{
		Instructions: `PicFast MCP Server — manage image hosting via AI.

Available tools:
- upload_image: Upload an image and get shareable links (URL, Markdown, HTML, BBCode).
- list_images: List your uploaded images with pagination.
- get_image: Get details and all format links for a specific image by key.
- delete_image: Delete an image by key.
- get_usage_stats: Get your storage usage and quota.

You can also read resources:
- picfast://user/profile — current user info and capacity.
- picfast://images — list of images (as resource).`,
	})

	// Register tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "upload_image",
		Description: "Upload an image to PicFast. Provide base64-encoded image data and a filename. Returns the image key, URL, and formatted links (Markdown, HTML, BBCode).",
	}, f.uploadImageTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_images",
		Description: "List uploaded images for the current user. Supports pagination.",
	}, f.listImagesTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_image",
		Description: "Get details and shareable links for a specific image by its key.",
	}, f.getImageTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_image",
		Description: "Delete an image by its key. This action is irreversible.",
	}, f.deleteImageTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_usage_stats",
		Description: "Get current user's storage usage statistics (used bytes, total capacity, image count).",
	}, f.getUsageStatsTool)

	// Register resources
	server.AddResource(&mcp.Resource{
		Name:     "user_profile",
		URI:      "picfast://user/profile",
		MIMEType: "application/json",
	}, f.userProfileResource)

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "image_detail",
		URITemplate: "picfast://images/{key}",
		MIMEType:    "application/json",
	}, f.imageDetailResource)

	// Register prompts
	server.AddPrompt(&mcp.Prompt{
		Name:        "upload_and_share",
		Description: "Upload an image and return all shareable formats",
	}, f.uploadAndSharePrompt)

	return server
}

// --- Tool handlers ---

type uploadImageArgs struct {
	FileData   string `json:"file_data"`
	Filename   string `json:"filename"`
	AlbumID    *int64 `json:"album_id,omitempty"`
	Permission *int16 `json:"permission,omitempty"`
}

func (f *MCPServerFactory) uploadImageTool(ctx context.Context, req *mcp.CallToolRequest, args uploadImageArgs) (*mcp.CallToolResult, any, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return errorResult("unauthorized: no user context"), nil, nil
	}

	fileData, err := base64.StdEncoding.DecodeString(args.FileData)
	if err != nil {
		return errorResult("invalid base64 file_data: " + err.Error()), nil, nil
	}

	uploadSvc := service.NewUploadService(f.DB, f.Pool, f.Config)
	result, err := uploadSvc.Store(ctx, service.UploadParams{
		FileData:   fileData,
		FileName:   args.Filename,
		FileSize:   int64(len(fileData)),
		AlbumID:    args.AlbumID,
		Permission: args.Permission,
		UserID:     &userID,
		ClientIP:   "mcp-client",
	})
	if err != nil {
		return errorResult("upload failed: " + err.Error()), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(`Image uploaded successfully!
Key: %s
URL: %s
Markdown: %s
HTML: %s
BBCode: %s`,
				result.Image.Key,
				result.Links.URL,
				result.Links.Markdown,
				result.Links.HTML,
				result.Links.BBCode,
			)},
		},
	}, nil, nil
}

type listImagesArgs struct {
	Page     int32 `json:"page,omitempty"`
	PageSize int32 `json:"page_size,omitempty"`
}

func (f *MCPServerFactory) listImagesTool(ctx context.Context, req *mcp.CallToolRequest, args listImagesArgs) (*mcp.CallToolResult, any, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return errorResult("unauthorized"), nil, nil
	}

	if args.Page <= 0 {
		args.Page = 1
	}
	if args.PageSize <= 0 || args.PageSize > 100 {
		args.PageSize = 20
	}

	images, err := f.DB.ListImagesByUser(ctx, sqlc.ListImagesByUserParams{
		UserID: domain.PgInt8(userID),
		Limit:  args.PageSize,
		Offset: (args.Page - 1) * args.PageSize,
	})
	if err != nil {
		return errorResult("failed to list images: " + err.Error()), nil, nil
	}

	total, _ := f.DB.CountImagesByUser(ctx, domain.PgInt8(userID))

	items := make([]map[string]any, len(images))
	for i, img := range images {
		items[i] = map[string]any{
			"key":        img.Key,
			"origin_name": img.OriginName,
			"size_bytes": img.SizeBytes,
			"width":      img.Width,
			"height":     img.Height,
			"url":        f.Config.Server.BaseURL + "/i/" + img.Key + "." + img.Extension,
			"permission": img.Permission,
			"created_at": img.CreatedAt,
		}
	}

	data, _ := json.Marshal(map[string]any{
		"items": items,
		"total": total,
		"page":  args.Page,
		"size":  args.PageSize,
	})

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

type getImageArgs struct {
	Key string `json:"key"`
}

func (f *MCPServerFactory) getImageTool(ctx context.Context, req *mcp.CallToolRequest, args getImageArgs) (*mcp.CallToolResult, any, error) {
	img, err := f.DB.GetImageByKey(ctx, args.Key)
	if err != nil {
		return errorResult("image not found"), nil, nil
	}

	url := f.Config.Server.BaseURL + "/i/" + img.Key + "." + img.Extension
	thumbURL := f.Config.Server.BaseURL + "/t/" + img.Md5 + ".png"

	data, _ := json.Marshal(map[string]any{
		"key":         img.Key,
		"origin_name": img.OriginName,
		"size_bytes":  img.SizeBytes,
		"mimetype":    img.Mimetype,
		"width":       img.Width,
		"height":      img.Height,
		"permission":  img.Permission,
		"url":         url,
		"thumbnail":   thumbURL,
		"markdown":    fmt.Sprintf("![%s](%s)", img.OriginName, url),
		"html":        fmt.Sprintf(`<img src="%s" alt="%s" />`, url, img.OriginName),
		"bbcode":      fmt.Sprintf("[img]%s[/img]", url),
		"created_at":  img.CreatedAt,
	})

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

type deleteImageArgs struct {
	Key string `json:"key"`
}

func (f *MCPServerFactory) deleteImageTool(ctx context.Context, req *mcp.CallToolRequest, args deleteImageArgs) (*mcp.CallToolResult, any, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return errorResult("unauthorized"), nil, nil
	}

	img, err := f.DB.GetImageByKey(ctx, args.Key)
	if err != nil {
		return errorResult("image not found"), nil, nil
	}

	// Permission check
	if img.UserID.Valid && img.UserID.Int64 != userID {
		return errorResult("permission denied: not your image"), nil, nil
	}

	deleteSvc := service.NewDeleteService(f.DB, f.Pool, f.Config.Storage.ThumbnailDir)
	if err := deleteSvc.DeleteImage(ctx, img.ID); err != nil {
		return errorResult("delete failed: " + err.Error()), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "Image deleted successfully."}},
	}, nil, nil
}

func (f *MCPServerFactory) getUsageStatsTool(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return errorResult("unauthorized"), nil, nil
	}

	user, err := f.DB.GetUserByID(ctx, userID)
	if err != nil {
		return errorResult("failed to get user info"), nil, nil
	}

	used, _ := f.DB.GetUserUsedCapacity(ctx, domain.PgInt8(userID))

	data, _ := json.Marshal(map[string]any{
		"user_id":        user.ID,
		"name":           user.Name,
		"capacity_bytes": user.CapacityBytes,
		"used_bytes":     used,
		"image_num":      user.ImageNum,
		"album_num":      user.AlbumNum,
	})

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

// --- Resources ---

func (f *MCPServerFactory) userProfileResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("unauthorized")
	}

	user, err := f.DB.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	used, _ := f.DB.GetUserUsedCapacity(ctx, domain.PgInt8(userID))

	data, _ := json.Marshal(map[string]any{
		"id":             user.ID,
		"name":           user.Name,
		"email":          user.Email,
		"role":           user.Role,
		"capacity_bytes": user.CapacityBytes,
		"used_bytes":     used,
		"image_num":      user.ImageNum,
		"album_num":      user.AlbumNum,
	})

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: "picfast://user/profile", MIMEType: "application/json", Text: string(data)},
		},
	}, nil
}

func (f *MCPServerFactory) imageDetailResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Extract key from URI picfast://images/{key}
	key := req.Params.URI[len("picfast://images/"):]
	if key == "" {
		return nil, fmt.Errorf("missing image key")
	}

	img, err := f.DB.GetImageByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("image not found")
	}

	url := f.Config.Server.BaseURL + "/i/" + img.Key + "." + img.Extension
	data, _ := json.Marshal(map[string]any{
		"key":         img.Key,
		"origin_name": img.OriginName,
		"size_bytes":  img.SizeBytes,
		"width":       img.Width,
		"height":      img.Height,
		"url":         url,
		"created_at":  img.CreatedAt,
	})

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: req.Params.URI, MIMEType: "application/json", Text: string(data)},
		},
	}, nil
}

// --- Prompts ---

func (f *MCPServerFactory) uploadAndSharePrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Description: "Upload an image and return all shareable link formats",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: "Please upload the image I provided to PicFast and give me the URL, Markdown, HTML, and BBCode formats.",
				},
			},
		},
	}, nil
}

// --- Helpers ---

func userIDFromContext(ctx context.Context) (int64, bool) {
	// MCP auth middleware stores TokenInfo in context
	tokenInfo := auth.TokenInfoFromContext(ctx)
	if tokenInfo == nil || tokenInfo.UserID == "" {
		return 0, false
	}
	userID, err := strconv.ParseInt(tokenInfo.UserID, 10, 64)
	if err != nil {
		return 0, false
	}
	return userID, true
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}
}
