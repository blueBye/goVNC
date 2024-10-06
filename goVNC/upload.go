package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/amitbet/vncproxy/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func upload(filename string, bucketName string) {
	ctx := context.Background()
	endpoint := os.Getenv("MINIO_ENDPPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		logger.Error("error creating client: ", err)
		return
	}

	// Make a new bucket called testbucket.
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			logger.Info("bucket exists")
		} else {
			logger.Error("error creating bucket: ", err)
			return
		}
	} else {
		logger.Info("bucket created")
	}

	// Upload the test file
	// Change the value of filePath if the file is in another location
	filePath := filename
	objectName := filepath.Base(filePath)
	contentType := "application/octet-stream"

	// Upload the test file with FPutObject
	info, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		logger.Error("failed to upload object:", err)
		return
	}
	logger.Info("upload succeeded:", objectName, "[", info.Size, "]")
}
