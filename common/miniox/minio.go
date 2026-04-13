package miniox

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

type Client struct {
	client *minio.Client
	bucket string
}

func NewClient(cfg Config) (*Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, err
	}
	if !exists {
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		client: client,
		bucket: cfg.BucketName,
	}, nil
}

func (c *Client) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := c.client.PutObject(ctx, c.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (c *Client) GetURL(objectName string, expires time.Duration) (string, error) {
	return c.client.PresignedGetObject(context.Background(), c.bucket, objectName, expires, nil)
}

func (c *Client) Delete(ctx context.Context, objectName string) error {
	return c.client.RemoveObject(ctx, c.bucket, objectName, minio.RemoveObjectOptions{})
}

func GenerateObjectName(userID int64, filename string) string {
	ext := filepath.Ext(filename)
	return fmt.Sprintf("uploads/%d/%d%s", userID, time.Now().UnixNano(), ext)
}
