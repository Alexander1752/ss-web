package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	storageOnce   sync.Once
	storageErr    error
	storageS3     *minio.Client
	storageBucket string
)

func initStorageClient(ctx context.Context) error {
	storageOnce.Do(func() {
		endpoint := strings.TrimSpace(os.Getenv("MINIO_ENDPOINT"))
		if endpoint == "" {
			storageErr = fmt.Errorf("MINIO_ENDPOINT is required")
			return
		}

		accessKey := os.Getenv("MINIO_ACCESS_KEY")
		secretKey := os.Getenv("MINIO_SECRET_KEY")

		useSSL := strings.EqualFold(strings.TrimSpace(os.Getenv("MINIO_USE_SSL")), "true")
		skipVerify := strings.EqualFold(strings.TrimSpace(os.Getenv("MINIO_INSECURE_SKIP_VERIFY")), "true")

		var transport http.RoundTripper
		if useSSL && skipVerify {
			transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}

		client, err := minio.New(endpoint, &minio.Options{
			Creds:     credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure:    useSSL,
			Transport: transport,
		})
		if err != nil {
			storageErr = fmt.Errorf("failed to initialize MinIO client: %w", err)
			return
		}

		storageS3 = client
		storageBucket = os.Getenv("MINIO_BUCKET_NAME")

		if storageBucket == "" {
			storageBucket = "local-bucket"
		}

		exists, err := storageS3.BucketExists(ctx, storageBucket)
		if err != nil {
			storageErr = fmt.Errorf("failed to check bucket %s: %w", storageBucket, err)
			return
		}
		if !exists {
			if err := storageS3.MakeBucket(ctx, storageBucket, minio.MakeBucketOptions{}); err != nil {
				storageErr = fmt.Errorf("failed to create bucket %s: %w", storageBucket, err)
				return
			}
		}
	})

	return storageErr
}

func SaveToMinIO(data []byte, keyName string) error {
	if err := initStorageClient(context.Background()); err != nil {
		return err
	}

	contentType := mime.TypeByExtension(path.Ext(keyName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := storageS3.PutObject(
		context.Background(),
		storageBucket,
		keyName,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return fmt.Errorf("failed to upload object %s: %w", keyName, err)
	}

	return nil
}

func GetPresignedURL(keyName string) string {
	if err := initStorageClient(context.Background()); err != nil {
		return ""
	}

	expiresIn := 3600
	if value := strings.TrimSpace(os.Getenv("MINIO_PRESIGNED_EXPIRY_SECONDS")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			expiresIn = parsed
		}
	}

	result, err := storageS3.PresignedGetObject(
		context.Background(),
		storageBucket,
		keyName,
		time.Duration(expiresIn)*time.Second,
		url.Values{},
	)
	if err != nil {
		return ""
	}

	return result.String()
}

func DeleteFromMinIO(keyName string) error {
	if err := initStorageClient(context.Background()); err != nil {
		return err
	}

	err := storageS3.RemoveObject(context.Background(), storageBucket, keyName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", keyName, err)
	}

	return nil
}

func DeletePrefixFromMinIO(prefix string) error {
	if err := initStorageClient(context.Background()); err != nil {
		return err
	}

	for obj := range storageS3.ListObjects(context.Background(), storageBucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
		if obj.Err != nil {
			return fmt.Errorf("failed to list objects for prefix %s: %w", prefix, obj.Err)
		}

		if err := DeleteFromMinIO(obj.Key); err != nil {
			return err
		}
	}

	return nil
}
