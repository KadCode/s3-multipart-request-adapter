package tests

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const testBucket = "test-bucket"

var s3Client *s3.Client

func TestMain(m *testing.M) {
	// Load AWS SDK config to connect to MinIO
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           "http://minio:9000", // MinIO container address in Docker Compose
					SigningRegion: "us-east-1",
				}, nil
			})),
	)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	s3Client = s3.NewFromConfig(cfg)

	// Create test bucket
	_, err = s3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(testBucket),
	})
	if err != nil {
		log.Fatalf("failed to create test bucket: %v", err)
	}
	log.Printf("Test bucket %s created", testBucket)

	// Run all tests
	code := m.Run()

	// Cleanup: delete all objects and the bucket itself
	objs, err := s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(testBucket),
	})
	if err == nil {
		for _, obj := range objs.Contents {
			_, _ = s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
				Bucket: aws.String(testBucket),
				Key:    obj.Key,
			})
		}
	}
	_, _ = s3Client.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket: aws.String(testBucket),
	})

	log.Printf("Test bucket %s deleted", testBucket)

	os.Exit(code)
}
