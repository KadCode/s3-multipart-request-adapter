package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
)

// ---------------------- UPLOAD ----------------------

// HandleCreateWithCtx uploads a file to S3 using a cancellable context
func HandleCreateWithCtx(ctx context.Context, s3Client *s3.Client, bucketName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		updateMaxMemory()
		start := time.Now()

		docID := c.Query("docId")
		if docID == "" {
			logRequest(c, start, "ERROR=missing docId")
			return c.Status(http.StatusBadRequest).SendString("docId required")
		}

		fileReader, filename, err := ExtractFileStream(c)
		if err != nil {
			logRequest(c, start, fmt.Sprintf("ERROR=%v", err))
			return c.Status(http.StatusBadRequest).SendString(fmt.Sprintf("file read error: %v", err))
		}

		contRep := c.Query("contRep")
		if contRep == "" {
			logRequest(c, start, "ERROR=missing contRep")
			return c.Status(http.StatusBadRequest).SendString("contRep required")
		}
		docID = strings.ToUpper(docID)

		rootDocTags := map[string]string{
			"contRep":  contRep,
			"docId":    docID,
			"filename": filename,
			"X-dateC":  time.Now().Format("2006-01-02"),
			"X-timeC":  time.Now().Format("15:04:05"),
			"X-dateM":  time.Now().Format("2006-01-02"),
			"X-timeM":  time.Now().Format("15:04:05"),
		}

		if filename == "" {
			filename = fmt.Sprintf("doc-%d", time.Now().Unix())
		}

		uploader := CreateS3Uploader(s3Client)

		// Upload using the cancellable context
		err = UploadFileToS3Stream(ctx, uploader, bucketName, docID, fileReader, rootDocTags)
		if err != nil {
			select {
			case <-ctx.Done():
				logRequest(c, start, "CANCELLED")
				return c.Status(fiber.StatusRequestTimeout).SendString("upload cancelled")
			default:
				logRequest(c, start, fmt.Sprintf("ERROR=%v", err))
				return c.Status(http.StatusInternalServerError).SendString(fmt.Sprintf("upload error: %v", err))
			}
		}

		logRequest(c, start, "UPLOADED")
		return c.Status(http.StatusOK).SendString(fmt.Sprintf("OK %s", filename))
	}
}

// ---------------------- DELETE ----------------------

// HandleDeleteWithCtx deletes a file from S3 using a cancellable context
func HandleDeleteWithCtx(ctx context.Context, s3Client *s3.Client, bucketName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		updateMaxMemory()
		start := time.Now()

		docID := c.Query("docId")
		if docID == "" {
			logRequest(c, start, "ERROR=missing docId")
			return c.Status(http.StatusBadRequest).SendString("docId required")
		}

		_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(docID),
		})
		if err != nil {
			select {
			case <-ctx.Done():
				logRequest(c, start, "CANCELLED")
				return c.Status(fiber.StatusRequestTimeout).SendString("delete cancelled")
			default:
				logRequest(c, start, fmt.Sprintf("ERROR=%v", err))
				return c.Status(http.StatusInternalServerError).SendString(fmt.Sprintf("delete error: %v", err))
			}
		}

		logRequest(c, start, "DELETED")
		return c.Status(http.StatusOK).SendString(fmt.Sprintf("DELETED %s", docID))
	}
}

// ---------------------- DOWNLOAD ----------------------

// HandleGetWithCtx downloads a file from S3 using a cancellable context
func HandleGetWithCtx(ctx context.Context, s3Client *s3.Client, bucketName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		updateMaxMemory()
		start := time.Now()

		docID := c.Query("docId")
		if docID == "" {
			logRequest(c, start, "ERROR=missing docId")
			return c.Status(http.StatusBadRequest).SendString("docId required")
		}

		// Read object metadata first
		head, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(docID),
		})
		if err != nil {
			select {
			case <-ctx.Done():
				logRequest(c, start, "CANCELLED")
				return c.Status(fiber.StatusRequestTimeout).SendString("request cancelled")
			default:
				logRequest(c, start, fmt.Sprintf("ERROR=%v", err))
				return c.Status(http.StatusNotFound).SendString(fmt.Sprintf("not found: %v", err))
			}
		}

		filename := docID
		if head.Metadata != nil && head.Metadata["filename"] != "" {
			filename = head.Metadata["filename"]
		}

		out, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(docID),
		})
		if err != nil {
			select {
			case <-ctx.Done():
				logRequest(c, start, "CANCELLED")
				return c.Status(fiber.StatusRequestTimeout).SendString("request cancelled")
			default:
				logRequest(c, start, fmt.Sprintf("ERROR=%v", err))
				return c.Status(http.StatusNotFound).SendString(fmt.Sprintf("not found: %v", err))
			}
		}
		defer out.Body.Close()

		c.Response().Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
		c.Status(http.StatusOK)

		n, err := io.Copy(c.Response().BodyWriter(), out.Body)
		if err != nil {
			select {
			case <-ctx.Done():
				logRequest(c, start, "CANCELLED")
				return c.Status(fiber.StatusRequestTimeout).SendString("download cancelled")
			default:
				logRequest(c, start, fmt.Sprintf("ERROR copying body: %v", err))
				return c.Status(http.StatusInternalServerError).SendString("internal error")
			}
		}

		logRequest(c, start, fmt.Sprintf("SERVED %s size=%d", filename, n))
		return nil
	}
}

// ---------------------- INFO ----------------------

// HandleInfoWithCtx provides metadata about an S3 object
func HandleInfoWithCtx(ctx context.Context, s3Client *s3.Client, bucketName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		updateMaxMemory()
		start := time.Now()

		docID := c.Query("docId")
		if docID == "" {
			logRequest(c, start, "ERROR=missing docId")
			return c.Status(http.StatusBadRequest).SendString("docId required")
		}

		head, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(docID),
		})
		if err != nil {
			select {
			case <-ctx.Done():
				logRequest(c, start, "CANCELLED")
				return c.Status(fiber.StatusRequestTimeout).SendString("request cancelled")
			default:
				logRequest(c, start, fmt.Sprintf("ERROR=%v", err))
				return c.Status(http.StatusNotFound).SendString(fmt.Sprintf("not found: %v", err))
			}
		}

		c.Status(fiber.StatusOK).JSON(fiber.Map{
			"docId":        docID,
			"size":         head.ContentLength,
			"lastModified": head.LastModified,
			"contentType":  head.ContentType,
			"etag":         head.ETag,
		})
		logRequest(c, start, "INFO")
		return nil
	}
}

// ---------------------- LIST ----------------------

// HandleListWithCtx lists objects in a bucket with context cancellation
func HandleListWithCtx(ctx context.Context, s3Client *s3.Client, bucketName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		updateMaxMemory()

		out, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			select {
			case <-ctx.Done():
				logRequest(c, start, "CANCELLED")
				return c.Status(fiber.StatusRequestTimeout).SendString("request cancelled")
			default:
				log.Printf("ListObjectsV2 error for bucket %s: %v", bucketName, err)
				logRequest(c, start, fmt.Sprintf("LIST count=0 ERROR=%v", err))
				return c.Status(fiber.StatusOK).JSON([]string{})
			}
		}

		items := make([]string, 0, len(out.Contents))
		for _, obj := range out.Contents {
			items = append(items, *obj.Key)
		}

		logRequest(c, start, fmt.Sprintf("LIST count=%d", len(items)))
		return c.Status(fiber.StatusOK).JSON(items)
	}
}

func HandleMem() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return c.Status(200).JSON(fiber.Map{
			"alloc":      m.Alloc,
			"totalAlloc": m.TotalAlloc,
			"sys":        m.Sys,
			"maxAlloc":   getMaxMemory(),
		})
	}
}

// HandleServerInfo returns runtime and memory info about the server
func HandleServerInfo() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		c.Status(fiber.StatusOK).JSON(fiber.Map{
			"goVersion":    runtime.Version(),
			"numCPU":       runtime.NumCPU(),
			"numGoroutine": runtime.NumGoroutine(),
			"alloc":        m.Alloc,
			"totalAlloc":   m.TotalAlloc,
			"sys":          m.Sys,
			"maxAlloc":     getMaxMemory(),
		})
		return nil
	}
}
