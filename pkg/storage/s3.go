package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct {
	Client           *s3.Client
	PresignClient    *s3.PresignClient
	BucketName       string
	Endpoint         string
	InternalEndpoint string
	accessKey        string
	secretKey        string
}

var S3 *S3Service

func InitS3(endpoint, region, bucket, accessKey, secretKey string, internalEndpoint string) error {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, reg string, options ...interface{}) (aws.Endpoint, error) {
		if endpoint != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           endpoint,
				SigningRegion: region,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config for S3: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true // Important for Garage and MinIO
	})

	presignClient := s3.NewPresignClient(client)

	S3 = &S3Service{
		Client:           client,
		PresignClient:    presignClient,
		BucketName:       bucket,
		Endpoint:         endpoint,
		InternalEndpoint: internalEndpoint,
		accessKey:        accessKey,
		secretKey:        secretKey,
	}

	log.Println("[OK]: S3 Service initialized successfully")
	return nil
}

func (s *S3Service) GeneratePresignedPutURL(ctx context.Context, objectKey string, lifetime time.Duration) (string, error) {
	// Always sign with the public endpoint so the Host header matches
	endpoint := s.Endpoint

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, reg string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           endpoint,
			SigningRegion: "us-east-1",
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s.accessKey, s.secretKey, "")),
	)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	presignClient := s3.NewPresignClient(client)

	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = lifetime
	})

	if err != nil {
		return "", fmt.Errorf("could not generate presigned URL: %w", err)
	}

	return req.URL, nil
}

// GetPublicURL returns the full URL for an object key (e.g. "items/uuid/image.png" -> "http://localhost:3900/nyx-items/items/uuid/image.png")
func (s *S3Service) GetPublicURL(objectKey string) string {
	return fmt.Sprintf("%s/%s/%s", s.Endpoint, s.BucketName, objectKey)
}
