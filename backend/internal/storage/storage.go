package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Storage defines the interface for audio storage
type Storage interface {
	Save(ctx context.Context, key string, data []byte) (string, error)
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	SaveAudio(ctx context.Context, key string, data []byte) error
	CloseChunk(ctx context.Context, key string) error
	GetSignedURL(ctx context.Context, key string) (string, error)
	Stream(ctx context.Context, key string) (io.ReadCloser, error)
}

// LocalStorage implements Storage interface for local filesystem
type LocalStorage struct {
	baseDir string
}

func NewLocalStorage(baseDir string) *LocalStorage {
	// Ensure base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		panic(err)
	}
	return &LocalStorage{baseDir: baseDir}
}

func (s *LocalStorage) Save(ctx context.Context, key string, data []byte) (string, error) {
	fullPath := filepath.Join(s.baseDir, key)
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", err
	}
	return key, nil
}

func (s *LocalStorage) Get(ctx context.Context, key string) ([]byte, error) {
	fullPath := filepath.Join(s.baseDir, key)
	return os.ReadFile(fullPath)
}

func (s *LocalStorage) CloseChunk(ctx context.Context, key string) error {
	// For local storage, no special handling needed
	return nil
}

// S3Storage implements Storage interface for AWS S3
type S3Storage struct {
	client *s3.Client
	bucket string
}

func NewS3Storage(accessKey, secretKey, region, bucket string) *S3Storage {
	creds := credentials.NewStaticCredentialsProvider(
		accessKey,
		secretKey,
		"",
	)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		panic(err)
	}

	client := s3.NewFromConfig(cfg)
	return &S3Storage{
		client: client,
		bucket: bucket,
	}
}

func (s *S3Storage) Save(ctx context.Context, key string, data []byte) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return "", err
	}
	return key, nil
}

func (s *S3Storage) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	return io.ReadAll(result.Body)
}

func (s *S3Storage) CloseChunk(ctx context.Context, key string) error {
	// For S3, we can add metadata to indicate the chunk is complete
	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		CopySource: aws.String(s.bucket + "/" + key),
		Key:        aws.String(key),
		Metadata: map[string]string{
			"status": "complete",
			"closed": time.Now().UTC().Format(time.RFC3339),
		},
		MetadataDirective: "REPLACE",
	})
	return err
}
