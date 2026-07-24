package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type uploadResponse struct {
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	Markdown     string `json:"markdown,omitempty"`
	BBCode       string `json:"bbcode,omitempty"`
	HTML         string `json:"html,omitempty"`
}

func main() {
	args := os.Args[1:]
	// Accept optional "upload" subcommand for npm README compatibility:
	//   picfast upload image.png
	if len(args) > 0 && args[0] == "upload" {
		args = args[1:]
	}

	fs := flag.NewFlagSet("picfast", flag.ExitOnError)
	markdown := fs.Bool("markdown", false, "output markdown image link instead of URL")
	format := fs.String("format", "", "output format: url, markdown, html, bbcode")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: picfast upload [flags] <file...>

  picfast upload image.png
  picfast upload --markdown image.png
  picfast upload *.png

Environment:
  PICFAST_URL    Base URL of your PicFast instance (required)
  PICFAST_TOKEN  API token for authenticated uploads (optional)

Flags:
`)
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	baseURL := strings.TrimRight(os.Getenv("PICFAST_URL"), "/")
	if baseURL == "" {
		fmt.Fprintln(os.Stderr, "error: PICFAST_URL is not set")
		os.Exit(1)
	}
	token := os.Getenv("PICFAST_TOKEN")

	files := fs.Args()
	if len(files) == 0 {
		fs.Usage()
		os.Exit(1)
	}

	exitCode := 0
	for _, path := range files {
		if err := uploadFile(baseURL, token, path, *markdown, *format); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s: %v\n", path, err)
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}

func uploadFile(baseURL, token, filePath string, markdown bool, format string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, f); err != nil {
		return err
	}
	w.Close()

	req, err := http.NewRequest("POST", baseURL+"/api/v1/flat/upload", &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("upload failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	var r uploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	out := r.URL
	if markdown || format == "markdown" {
		out = r.Markdown
	} else if format == "html" {
		out = r.HTML
	} else if format == "bbcode" {
		out = r.BBCode
	}
	fmt.Println(out)
	return nil
}
