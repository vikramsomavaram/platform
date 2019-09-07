/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package storage

import (
	"cloud.google.com/go/storage"
	"context"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	// StorageBucket ...
	StorageBucket *storage.BucketHandle
	//StorageBucketName ...
	StorageBucketName string
)

func init() {
	StorageBucketName = os.Getenv("GCS_BUCKET")
	storageBucket, err := configureStorage(StorageBucketName)
	if err != nil {
		log.Fatal(err)
	}
	StorageBucket = storageBucket
}

func configureStorage(bucketID string) (*storage.BucketHandle, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Bucket(bucketID), nil
}
