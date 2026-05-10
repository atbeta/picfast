package maintenance

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type InspectResult struct {
	GeneratedAt      time.Time `json:"generated_at"`
	ArchivePath      string    `json:"archive_path"`
	Manifest         Manifest  `json:"manifest"`
	ChecksumCount    int       `json:"checksum_count"`
	VerifiedCount    int       `json:"verified_count"`
	MissingChecksums []string  `json:"missing_checksums,omitempty"`
	ChecksumFailures []string  `json:"checksum_failures,omitempty"`
	Warnings         []string  `json:"warnings,omitempty"`
}

func (r InspectResult) OK() bool {
	return len(r.MissingChecksums) == 0 && len(r.ChecksumFailures) == 0
}

func InspectArchive(path string) (InspectResult, error) {
	result := InspectResult{GeneratedAt: time.Now().UTC(), ArchivePath: path}
	file, err := os.Open(path)
	if err != nil {
		return result, err
	}
	defer file.Close()

	reader, closeFn, err := archiveReader(path, file)
	if err != nil {
		return result, err
	}
	if closeFn != nil {
		defer closeFn()
	}

	actual := map[string]ChecksumEntry{}
	var checksums []ChecksumEntry
	var objects []ObjectManifestEntry
	var sawManifest bool
	var sawChecksums bool

	tr := tar.NewReader(reader)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return result, err
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		hash := sha256.New()
		data, err := io.ReadAll(io.TeeReader(tr, hash))
		if err != nil {
			return result, err
		}
		actual[header.Name] = ChecksumEntry{
			Path:      header.Name,
			SizeBytes: int64(len(data)),
			SHA256:    hex.EncodeToString(hash.Sum(nil)),
		}
		switch header.Name {
		case "manifest.json":
			if err := json.Unmarshal(data, &result.Manifest); err != nil {
				return result, fmt.Errorf("decode manifest: %w", err)
			}
			sawManifest = true
		case "checksums.jsonl":
			parsed, err := parseChecksumJSONL(strings.NewReader(string(data)))
			if err != nil {
				return result, err
			}
			checksums = parsed
			sawChecksums = true
		case "objects.jsonl":
			parsed, err := parseObjectManifestJSONL(strings.NewReader(string(data)))
			if err != nil {
				return result, err
			}
			objects = parsed
		}
	}

	if !sawManifest {
		return result, fmt.Errorf("manifest.json not found")
	}
	if err := result.Manifest.Validate(); err != nil {
		return result, err
	}
	if !sawChecksums {
		return result, fmt.Errorf("checksums.jsonl not found")
	}

	result.ChecksumCount = len(checksums)
	expectedChecksumPaths := make(map[string]bool, len(checksums))
	objectArchivePaths := make(map[string]bool, len(objects))
	for _, object := range objects {
		if object.ArchivePath != "" {
			objectArchivePaths[object.ArchivePath] = true
		}
	}
	for _, expected := range checksums {
		expectedChecksumPaths[expected.Path] = true
		if strings.HasPrefix(expected.Path, "objects/") && !objectArchivePaths[expected.Path] {
			result.ChecksumFailures = append(result.ChecksumFailures, expected.Path)
			continue
		}
		got, ok := actual[expected.Path]
		if !ok {
			result.MissingChecksums = append(result.MissingChecksums, expected.Path)
			continue
		}
		if got.SizeBytes != expected.SizeBytes || got.SHA256 != expected.SHA256 {
			result.ChecksumFailures = append(result.ChecksumFailures, expected.Path)
			continue
		}
		result.VerifiedCount++
	}
	for _, object := range objects {
		if object.ArchivePath == "" {
			continue
		}
		if !expectedChecksumPaths[object.ArchivePath] {
			result.MissingChecksums = append(result.MissingChecksums, object.ArchivePath)
		}
	}
	return result, nil
}

func archiveReader(path string, file *os.File) (io.Reader, func() error, error) {
	if strings.HasSuffix(path, ".gz") || strings.HasSuffix(path, ".tgz") {
		gz, err := gzip.NewReader(file)
		if err != nil {
			return nil, nil, err
		}
		return gz, gz.Close, nil
	}
	return file, nil, nil
}

func parseChecksumJSONL(r io.Reader) ([]ChecksumEntry, error) {
	var checksums []ChecksumEntry
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var checksum ChecksumEntry
		if err := json.Unmarshal([]byte(line), &checksum); err != nil {
			return nil, fmt.Errorf("decode checksums.jsonl: %w", err)
		}
		checksums = append(checksums, checksum)
	}
	return checksums, scanner.Err()
}

func parseObjectManifestJSONL(r io.Reader) ([]ObjectManifestEntry, error) {
	var objects []ObjectManifestEntry
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var object ObjectManifestEntry
		if err := json.Unmarshal([]byte(line), &object); err != nil {
			return nil, fmt.Errorf("decode objects.jsonl: %w", err)
		}
		objects = append(objects, object)
	}
	return objects, scanner.Err()
}
