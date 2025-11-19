package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	s3_adapter_config "example.com/s3-multipart-request-adapter/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
)

// CreateS3Client initializes an S3 client using configuration
func CreateS3Client() *s3.Client {
	cfg, err := s3_adapter_config.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost: cfg.S3.MaxConnections,
		},
	}

	// AWS SDK Config
	s3Cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(cfg.S3.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.S3.AccessKey, cfg.S3.SecretKey, ""),
		),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL:               cfg.S3.Url,
						SigningRegion:     cfg.S3.Region,
						HostnameImmutable: true,
					}, nil
				},
			),
		),
		config.WithHTTPClient(httpClient),
	)
	if err != nil {
		log.Fatalf("AWS config load error: %v", err)
	}

	return s3.NewFromConfig(s3Cfg)
}

// EncodeTags converts a key-value map into a query string for S3 object tagging
func EncodeTags(tags map[string]string) string {
	var sb strings.Builder
	i := 0
	for k, v := range tags {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(v)
		i++
	}
	return sb.String()
}

// UploadFileToS3Stream uploads a file stream to S3
func UploadFileToS3Stream(ctx context.Context, uploader *manager.Uploader, bucketName, key string, body io.Reader, tags map[string]string) error {
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:  aws.String(bucketName),
		Key:     aws.String(key),
		Body:    body,
		Tagging: aws.String(EncodeTags(tags)),
	})
	return err
}

// CreateS3Uploader creates a high-level S3 uploader
func CreateS3Uploader(client *s3.Client) *manager.Uploader {
	return manager.NewUploader(client)
}

// ExtractFileStream extracts the first file stream from a multipart form
func ExtractFileStream(c *fiber.Ctx) (io.Reader, string, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, "", fmt.Errorf("multipart reader: %w", err)
	}

	for _, files := range form.File {
		if len(files) == 0 {
			continue
		}

		file := files[0]
		f, err := file.Open()
		if err != nil {
			return nil, "", fmt.Errorf("open file: %w", err)
		}

		// IMPORTANT: do not close f here, Fiber will handle closing
		return f, file.Filename, nil
	}

	return nil, "", fmt.Errorf("no file part found")
}
