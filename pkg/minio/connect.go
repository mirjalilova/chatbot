package minio

import (
	"context"
	"fmt"

	"chatbot/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"golang.org/x/exp/slog"
)

type MinIO struct {
	Client *minio.Client
	Cnf    *config.Config
}

var bucketName = "ai1009"

func MinIOConnect(cnf *config.Config) (*MinIO, error) {

	minioClient, err := minio.New(cnf.MINIO_ENDPOINT, &minio.Options{
		Creds:  credentials.NewStaticV4(cnf.MINIO_ACCESS_KEY, cnf.MINIO_SECRET_KEY, ""),
		Secure: false,
	})
	if err != nil {
		slog.Error("Failed to connect to MinIO: %v", err)
		return nil, err
	}

	err = minioClient.MakeBucket(context.Background(), cnf.MINIO_BUCKET_NAME, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(context.Background(), cnf.MINIO_BUCKET_NAME)
		if errBucketExists == nil && exists {
			slog.Warn("Bucket already exists: %s\n", cnf.MINIO_BUCKET_NAME)
		} else {
			slog.Error("Error while making bucket %s: %v\n", cnf.MINIO_BUCKET_NAME, err)
		}
	} else {
		slog.Info("Successfully created bucket: %s\n", cnf.MINIO_BUCKET_NAME)
	}

	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": "*",
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}
		]
	}`, cnf.MINIO_BUCKET_NAME)

	err = minioClient.SetBucketPolicy(context.Background(), cnf.MINIO_BUCKET_NAME, policy)
	if err != nil {
		slog.Error("Error while setting bucket policy: %v", err)
		return nil, err
	}

	return &MinIO{
		Client: minioClient,
		Cnf:    cnf,
	}, nil
}

func (m *MinIO) Upload(fileName, filePath string) (string, error) {
	_, err := m.Client.FPutObject(
		context.Background(),
		m.Cnf.MINIO_BUCKET_NAME,
		fileName,
		filePath,
		minio.PutObjectOptions{},
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"%s/%s/%s",
		m.Cnf.MINIO_PUBLIC_URL,
		m.Cnf.MINIO_BUCKET_NAME,
		fileName,
	), nil
}

