package storage

import (
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client  *minio.Client
	options *Options
}

type Options struct {
	Endpoint string   `json:"endpoint"`
	ID       string   `json:"id"`
	Secret   string   `json:"secret"`
	Buckets  []string `json:"buckets"`
}

func NewClient(options *Options) *Client {
	client, err := minio.New(options.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(options.ID, options.Secret, ""),
		Secure: true,
	})
	if err != nil {
		panic(err)
	}

	for _, bucket := range options.Buckets {
		if err = client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
			if exists, errBucketExists := client.BucketExists(context.Background(), bucket); !exists || errBucketExists != nil {
				log.Fatalln(errBucketExists)
			}
		}
	}

	return &Client{
		client:  client,
		options: options,
	}
}

func (c *Client) upload() {

}
