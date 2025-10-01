package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

var (
	versionPattern = regexp.MustCompile(`\b\d{2}\.\d{2}\.\d{2}\.\d+\b`)
	targetFiles    = []string{
		"doc.go",
		"cmd/server/main.go",
		"internal/app.go",
		"internal/swagger/docs/docs.go",
		"internal/swagger/docs/swagger.json",
		"internal/swagger/docs/swagger.yaml",
	}
)

func main() {
	log.SetFlags(0)

	repo := flag.String("repo", ".", "path to repository root")
	version := flag.String("version", "", "version string to stamp into swagger metadata")
	flag.Parse()

	if *version == "" {
		log.Fatal("missing required --version value")
	}

	if !versionPattern.MatchString(*version) {
		log.Fatalf("version %q does not match expected YY.MM.DD.B format", *version)
	}

	root, err := filepath.Abs(*repo)
	if err != nil {
		log.Fatalf("resolve repo path: %v", err)
	}

	for _, rel := range targetFiles {
		if err := stampVersion(root, rel, *version); err != nil {
			log.Fatal(err)
		}
	}
}

func stampVersion(repo, rel, version string) error {
	path := filepath.Join(repo, rel)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", rel, err)
	}

	if !versionPattern.Match(data) {
		return fmt.Errorf("no version token found in %s", rel)
	}

	updated := versionPattern.ReplaceAll(data, []byte(version))
	if bytes.Equal(data, updated) {
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %s: %w", rel, err)
	}

	if err := os.WriteFile(path, updated, info.Mode().Perm()); err != nil {
		return fmt.Errorf("write %s: %w", rel, err)
	}

	return nil
}
