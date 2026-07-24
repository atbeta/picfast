package main

import (
	"bufio"
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
	if len(args) == 0 {
		printRootUsage(os.Stderr)
		os.Exit(1)
	}

	switch args[0] {
	case "config":
		os.Exit(runConfigCommand(args[1:]))
	case "help", "-h", "--help":
		printRootUsage(os.Stdout)
		os.Exit(0)
	case "upload":
		os.Exit(runUploadCommand(args[1:]))
	default:
		// Bare files: picfast image.png
		os.Exit(runUploadCommand(args))
	}
}

func printRootUsage(w io.Writer) {
	fmt.Fprintf(w, `Usage:
  picfast upload [flags] <file...>
  picfast config set url <url>
  picfast config set token [token]
  picfast config set token --stdin
  picfast config show
  picfast config unset <url|token>

Environment (overrides config file):
  PICFAST_URL         Base URL of your PicFast instance
  PICFAST_TOKEN       API token for authenticated uploads
  PICFAST_CONFIG_DIR  Override config directory

Config file:
  <user-config-dir>/picfast/config.json
`)
}

func runUploadCommand(args []string) int {
	fs := flag.NewFlagSet("picfast upload", flag.ExitOnError)
	markdown := fs.Bool("markdown", false, "output markdown image link instead of URL")
	format := fs.String("format", "", "output format: url, markdown, html, bbcode")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: picfast upload [flags] <file...>

  picfast upload image.png
  picfast upload --markdown image.png
  picfast upload *.png

Credentials:
  PICFAST_URL / PICFAST_TOKEN environment variables override
  values from: picfast config set ...

Flags:
`)
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	baseURL, token, err := resolveCredentials()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read config: %v\n", err)
		return 1
	}
	if baseURL == "" {
		fmt.Fprintln(os.Stderr, "error: PICFAST_URL is not set (use env or: picfast config set url <url>)")
		return 1
	}

	files := fs.Args()
	if len(files) == 0 {
		fs.Usage()
		return 1
	}

	exitCode := 0
	for _, path := range files {
		if err := uploadFile(baseURL, token, path, *markdown, *format); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s: %v\n", path, err)
			exitCode = 1
		}
	}
	return exitCode
}

func runConfigCommand(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: picfast config <set|show|unset> ...")
		return 1
	}
	switch args[0] {
	case "show":
		return runConfigShow()
	case "set":
		return runConfigSet(args[1:])
	case "unset":
		return runConfigUnset(args[1:])
	case "help", "-h", "--help":
		printRootUsage(os.Stdout)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "error: unknown config command %q\n", args[0])
		return 1
	}
}

func runConfigShow() int {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	path, err := configPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	url := cfg.URL
	if url == "" {
		url = "(not set)"
	}
	fmt.Printf("config: %s\n", path)
	fmt.Printf("url:    %s\n", url)
	fmt.Printf("token:  %s\n", maskToken(cfg.Token))
	if envURL := strings.TrimSpace(os.Getenv("PICFAST_URL")); envURL != "" {
		fmt.Printf("note:   PICFAST_URL env is set and overrides file url\n")
	}
	if envTok := strings.TrimSpace(os.Getenv("PICFAST_TOKEN")); envTok != "" {
		fmt.Printf("note:   PICFAST_TOKEN env is set and overrides file token\n")
	}
	return 0
}

func runConfigSet(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: picfast config set <url|token> [value]")
		return 1
	}
	key := args[0]
	rest := args[1:]

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	switch key {
	case "url":
		if len(rest) != 1 || strings.TrimSpace(rest[0]) == "" {
			fmt.Fprintln(os.Stderr, "usage: picfast config set url <url>")
			return 1
		}
		cfg.URL = rest[0]
	case "token":
		value, err := readTokenValue(rest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		if value == "" {
			fmt.Fprintln(os.Stderr, "error: token value is empty")
			return 1
		}
		cfg.Token = value
	default:
		fmt.Fprintf(os.Stderr, "error: unknown config key %q (want url or token)\n", key)
		return 1
	}

	if err := saveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: save config: %v\n", err)
		return 1
	}
	fmt.Printf("saved %s\n", key)
	return 0
}

func readTokenValue(args []string) (string, error) {
	stdin := false
	var positional []string
	for _, a := range args {
		if a == "--stdin" {
			stdin = true
			continue
		}
		positional = append(positional, a)
	}
	if stdin {
		if len(positional) > 0 {
			return "", fmt.Errorf("token: use either --stdin or a value, not both")
		}
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	}
	if len(positional) == 0 {
		fmt.Fprint(os.Stderr, "Enter API token: ")
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && err != io.EOF {
			return "", err
		}
		return strings.TrimSpace(line), nil
	}
	if len(positional) > 1 {
		return "", fmt.Errorf("usage: picfast config set token [token] | --stdin")
	}
	return strings.TrimSpace(positional[0]), nil
}

func runConfigUnset(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: picfast config unset <url|token>")
		return 1
	}
	if err := unsetConfigKey(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Printf("unset %s\n", args[0])
	return 0
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
