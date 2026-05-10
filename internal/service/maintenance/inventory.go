package maintenance

import (
	"context"
	"database/sql"
	"encoding/json"
	"path"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type InventoryOptions struct {
	Limit  int32
	Offset int32
}

type InventoryItem struct {
	ImageID        int64           `json:"image_id"`
	Key            string          `json:"key"`
	StrategyID     *int64          `json:"strategy_id,omitempty"`
	StrategyName   string          `json:"strategy_name,omitempty"`
	StrategyType   string          `json:"strategy_type,omitempty"`
	StrategyConfig json.RawMessage `json:"-"`
	ObjectPath     string          `json:"object_path"`
	SizeBytes      int64           `json:"size_bytes"`
	MimeType       string          `json:"mimetype"`
	Extension      string          `json:"extension"`
	MD5            string          `json:"md5"`
	SHA1           string          `json:"sha1"`
	CreatedAt      time.Time       `json:"created_at"`
}

func (item InventoryItem) ThumbnailName() string {
	if item.Extension == "svg" || item.Extension == "ico" || item.MD5 == "" {
		return ""
	}
	return item.MD5 + ".png"
}

type InventorySource interface {
	ListImages(ctx context.Context, opts InventoryOptions) ([]InventoryItem, error)
}

type PGInventorySource struct {
	pool *pgxpool.Pool
}

func NewPGInventorySource(pool *pgxpool.Pool) *PGInventorySource {
	return &PGInventorySource{pool: pool}
}

func (s *PGInventorySource) ListImages(ctx context.Context, opts InventoryOptions) ([]InventoryItem, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 500
	}
	if limit > 5000 {
		limit = 5000
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			images.id,
			images.key,
			images.strategy_id,
			COALESCE(strategies.name, ''),
			COALESCE(strategies.strategy_type, ''),
			COALESCE(strategies.configs, '{}'::jsonb),
			images.path,
			images.name,
			images.size_bytes,
			images.mimetype,
			images.extension,
			images.md5,
			images.sha1,
			images.created_at
		FROM images
		LEFT JOIN strategies ON images.strategy_id = strategies.id
		ORDER BY images.id ASC
		LIMIT $1 OFFSET $2
	`, limit, opts.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]InventoryItem, 0, limit)
	for rows.Next() {
		var item InventoryItem
		var strategyID sql.NullInt64
		var imagePath, imageName string
		if err := rows.Scan(
			&item.ImageID,
			&item.Key,
			&strategyID,
			&item.StrategyName,
			&item.StrategyType,
			&item.StrategyConfig,
			&imagePath,
			&imageName,
			&item.SizeBytes,
			&item.MimeType,
			&item.Extension,
			&item.MD5,
			&item.SHA1,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if strategyID.Valid {
			item.StrategyID = &strategyID.Int64
		}
		item.ObjectPath = joinObjectPath(imagePath, imageName)
		items = append(items, item)
	}
	return items, rows.Err()
}

func joinObjectPath(dir, name string) string {
	dir = strings.Trim(dir, "/")
	name = strings.TrimLeft(name, "/")
	if dir == "" {
		return name
	}
	return path.Join(dir, name)
}
