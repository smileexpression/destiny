package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"

	"smile.expression/destiny/pkg/constant"
	"smile.expression/destiny/pkg/http/api"
	"smile.expression/destiny/pkg/logger"
)

type Client struct {
	client  *minio.Client
	options *Options
}

type Options struct {
	Endpoint string   `json:"endpoint"`
	ID       string   `json:"id"`
	Secret   string   `json:"secret"`
	Secure   bool     `json:"secure"`
	Buckets  []string `json:"buckets"`
}

func NewClient(options *Options) *Client {
	var (
		log = logger.SmileLog.Logger
	)

	client, err := minio.New(options.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(options.ID, options.Secret, ""),
		Secure: options.Secure,
	})
	if err != nil {
		log.WithError(err).Fatal("error creating new client")
		panic(err)
	}

	for _, bucket := range options.Buckets {
		if err = client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
			if exists, errBucketExists := client.BucketExists(context.Background(), bucket); !exists || errBucketExists != nil {
				log.WithError(err).Errorf("Failed to create bucket: %s", bucket)
				panic(errBucketExists)
			}
		}
	}

	return &Client{
		client:  client,
		options: options,
	}
}

func (c *Client) PutObject(ctx context.Context, bucketName string, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (*api.PutObjectResponse, error) {
	var (
		log = logger.SmileLog.WithContext(ctx).WithFields(logrus.Fields{
			constant.Route:  constant.PutObject,
			constant.Bucket: bucketName,
			constant.Object: objectName,
			constant.Size:   objectSize,
		})
	)

	info, err := c.client.PutObject(ctx, bucketName, objectName, reader, objectSize, opts)
	if err != nil {
		log.WithError(err).Error("PutObject fail")
		return nil, err
	}
	log.Info("PutObject success")

	respURL := fmt.Sprintf("http://%s/%s/%s", c.options.Endpoint, bucketName, objectName)
	return &api.PutObjectResponse{
		URL:  respURL,
		ETag: info.ETag,
		Size: info.Size,
	}, nil
}

func (c *Client) RemoveObject(ctx context.Context, bucketName string, objectName string, opts minio.RemoveObjectOptions) error {
	var (
		log = logger.SmileLog.WithContext(ctx).WithFields(logrus.Fields{
			constant.Route:  constant.RemoveObject,
			constant.Bucket: bucketName,
			constant.Object: objectName,
		})
	)

	if err := c.client.RemoveObject(ctx, bucketName, objectName, opts); err != nil {
		log.WithError(err).Error("RemoveObject fail")
		return err
	}
	log.Info("RemoveObject success")

	return nil
}

func (c *Client) SetEndpoint(uri string) string {
	return fmt.Sprintf("http://%s/%s", c.options.Endpoint, uri)
}

func (c *Client) GetObjectName(url string) string {
	_, objectName, err := c.ParseURL(url)
	if err != nil {
		return ""
	}
	return objectName
}

func (c *Client) ParseURL(fileURL string) (string, string, error) {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", "", err
	}

	// 解析路径，例如 "/my-bucket/my-file.txt"
	path := strings.TrimPrefix(parsedURL.Path, "/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid URL format")
	}

	return parts[0], parts[1], nil
}
