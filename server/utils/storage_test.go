package utils

import (
	"strings"
	"sync"
	"testing"
)

func resetStorageState(t *testing.T) {
	t.Helper()
	storageOnce = sync.Once{}
	storageErr = nil
	storageS3 = nil
	storageBucket = ""
}

func TestStorageOperationsRequireMinIOEndpoint(t *testing.T) {
	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "save",
			run: func() error {
				return SaveToMinIO([]byte("image-bytes"), "photos/123.png")
			},
		},
		{
			name: "delete object",
			run: func() error {
				return DeleteFromMinIO("photos/123.png")
			},
		},
		{
			name: "delete prefix",
			run: func() error {
				return DeletePrefixFromMinIO("photos/")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetStorageState(t)
			t.Setenv("MINIO_ENDPOINT", "")

			err := tt.run()
			if err == nil {
				t.Fatalf("expected missing endpoint error")
			}
			if !strings.Contains(err.Error(), "MINIO_ENDPOINT is required") {
				t.Fatalf("expected missing endpoint error, got %v", err)
			}
		})
	}
}

func TestGetPresignedURL_ReturnsEmptyWhenStorageCannotInitialize(t *testing.T) {
	resetStorageState(t)
	t.Setenv("MINIO_ENDPOINT", "")

	if url := GetPresignedURL("photos/123.png"); url != "" {
		t.Fatalf("expected empty URL when storage initialization fails, got %q", url)
	}
}
