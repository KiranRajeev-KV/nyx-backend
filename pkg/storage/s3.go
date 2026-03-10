package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct {
	Client        *s3.Client
	PresignClient *s3.PresignClient
	BucketName    string
	Endpoint      string
	accessKey     string
	secretKey     string
}

var S3 *S3Service

func InitS3(endpoint, region, bucket, accessKey, secretKey string) error {
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
		o.UsePathStyle = true // Important for Garage, MinIO, and Supabase
	})

	presignClient := s3.NewPresignClient(client)

	S3 = &S3Service{
		Client:        client,
		PresignClient: presignClient,
		BucketName:    bucket,
		Endpoint:      endpoint,
		accessKey:     accessKey,
		secretKey:     secretKey,
	}

	// log.Println("[OK]: S3 Service initialized successfully")
	return nil
}

func (s *S3Service) GeneratePresignedPutURL(ctx context.Context, objectKey string, lifetime time.Duration) (string, error) {
	req, err := s.PresignClient.PresignPutObject(ctx, &s3.PutObjectInput{
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

func (s *S3Service) GeneratePresignedGetURL(ctx context.Context, objectKey string, lifetime time.Duration) (string, error) {
	req, err := s.PresignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = lifetime
	})

	if err != nil {
		return "", fmt.Errorf("could not generate presigned GET URL: %w", err)
	}

	return req.URL, nil
}

// GetPublicURL returns the full public URL for an object key.
// For Supabase: https://[project-ref].supabase.co/storage/v1/object/public/[bucket]/[key]
// For S3/Garage: https://[endpoint]/[bucket]/[key]
func (s *S3Service) GetPublicURL(objectKey string) string {
	// Check if this is Supabase (endpoint contains .storage.supabase.co)
	if strings.Contains(s.Endpoint, ".storage.supabase.co") {
		// Extract project-ref from endpoint
		// e.g., https://xxx.storage.supabase.co/storage/v1/s3 -> xxx.supabase.co
		projectRef := strings.ReplaceAll(s.Endpoint, ".storage.supabase.co/storage/v1/s3", "")
		projectRef = strings.TrimPrefix(projectRef, "https://")
		return fmt.Sprintf("https://%s.supabase.co/storage/v1/object/public/%s/%s", projectRef, s.BucketName, objectKey)
	}

	// Default S3/Garage format
	return fmt.Sprintf("%s/%s/%s", s.Endpoint, s.BucketName, objectKey)
}
