package tests

import (
	"context"
	"log"
	"os"
	"testing"

	testconfig "example.com/s3-multipart-request-adapter/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const testBucket = "test-bucket"

var s3Client *s3.Client

func TestMain(m *testing.M) {
	testCfg, err := testconfig.GetTestConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	// Load AWS SDK config to connect to MinIO
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(testCfg.S3.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			testCfg.S3.AccessKey, testCfg.S3.SecretKey, testCfg.S3.Session)),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               testCfg.S3.Url,
					SigningRegion:     testCfg.S3.Region,
					HostnameImmutable: testCfg.S3.HostnameImmutable,
				}, nil
			})),
	)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	_, err = s3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(testBucket),
	})
	if err != nil {
		log.Fatalf("failed to create test bucket: %v", err)
	}
	log.Printf("Test bucket %s created", testBucket)

	code := m.Run()

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
