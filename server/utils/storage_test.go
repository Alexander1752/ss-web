package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveToLocal_CreatesDirectoriesAndWritesFile(t *testing.T) {
	tempDir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to chdir temp: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWD)
	})

	content := []byte("image-bytes")
	err = SaveToLocal(content, filepath.Join("photos", "123.png"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	path := filepath.Join(tempDir, "uploads", "photos", "123.png")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected file to be created, got %v", err)
	}
	if string(data) != string(content) {
		t.Fatalf("expected written content %q, got %q", string(content), string(data))
	}
}

func TestGetLocalURL_UsesDefaultAndEnvBaseURL(t *testing.T) {
	t.Run("uses default base url", func(t *testing.T) {
		t.Setenv("API_BASE_URL", "")
		url := GetLocalURL("photos/1.png")
		if url != "http://localhost:8080/uploads/photos/1.png" {
			t.Fatalf("unexpected default URL: %s", url)
		}
	})

	t.Run("trims trailing slash from env base url", func(t *testing.T) {
		t.Setenv("API_BASE_URL", "https://api.example.com/")
		url := GetLocalURL("photos/2.png")
		if url != "https://api.example.com/uploads/photos/2.png" {
			t.Fatalf("unexpected env URL: %s", url)
		}
		if strings.Contains(url, "//uploads") {
			t.Fatalf("expected normalized URL without double slash, got %s", url)
		}
	})
}
