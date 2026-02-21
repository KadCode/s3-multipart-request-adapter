package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"example.com/s3-multipart-request-adapter/utils"

	_ "example.com/s3-multipart-request-adapter/docs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/google/uuid"
)

var (
	activeRequests int32    // counter of currently active requests
	requestCtxs    sync.Map // map to store cancel functions for active requests
)

func main() {
	// Create S3 client
	var s3Client = utils.CreateS3Client()

	//Fiber configuration
	app := utils.CreateNewFiberAppInstance()

	contentRoute := "/ContentServer/ContentServer.dll"

	// Middleware to track active requests and create a cancellable context
	app.Use(func(c *fiber.Ctx) error {
		requestID := uuid.New().String()
		c.Locals("requestID", requestID) //  generate a UUID for each request
		c.Set("X-Request-ID", requestID)
		atomic.AddInt32(&activeRequests, 1)        // increment active request count
		defer atomic.AddInt32(&activeRequests, -1) // decrement when finished

		// Create a context with cancel for this request
		ctx, cancel := context.WithCancel(context.Background())
		requestCtxs.Store(requestID, cancel) // store cancel function
		defer requestCtxs.Delete(requestID)  // remove from map when done

		c.Locals("ctx", ctx) // attach context to request
		return c.Next()
	})

	// Routes
	app.Get("/mem", utils.HandleMem())
	app.Get("/docs/*", swagger.HandlerDefault)

	app.Get(contentRoute, func(c *fiber.Ctx) error {
		ctx := c.Locals("ctx").(context.Context)
		bucketName := c.Query("contRep")
		if bucketName == "" {
			return c.Status(fiber.StatusBadRequest).SendString("missing contRep")
		}

		q := c.Queries()
		_, isGet := q["get"]
		_, isInfo := q["info"]
		_, isList := q["list"]
		_, isServerInfo := q["serverInfo"]

		if (isGet || isInfo) && c.Query("docId") == "" {
			return c.Status(fiber.StatusBadRequest).SendString("missing docId")
		}

		switch {
		case isGet:
			return utils.HandleGetWithCtx(ctx, s3Client, bucketName)(c)
		case isInfo:
			return utils.HandleInfoWithCtx(ctx, s3Client, bucketName)(c)
		case isList:
			return utils.HandleListWithCtx(ctx, s3Client, bucketName)(c)
		case isServerInfo:
			return utils.HandleServerInfo()(c)
		default:
			return c.Status(fiber.StatusBadRequest).SendString("unknown action")
		}
	})

	app.Post(contentRoute, func(c *fiber.Ctx) error {
		ctx := c.Locals("ctx").(context.Context)
		bucketName := c.Query("contRep")
		if bucketName == "" {
			return c.Status(fiber.StatusBadRequest).SendString("missing contRep")
		}
		if c.Query("docId") == "" {
			return c.Status(fiber.StatusBadRequest).SendString("missing docId")
		}
		return utils.HandleCreateWithCtx(ctx, s3Client, bucketName)(c)
	})

	app.Delete(contentRoute, func(c *fiber.Ctx) error {
		ctx := c.Locals("ctx").(context.Context)
		bucketName := c.Query("contRep")
		if bucketName == "" {
			return c.Status(fiber.StatusBadRequest).SendString("missing contRep")
		}
		if c.Query("docId") == "" {
			return c.Status(fiber.StatusBadRequest).SendString("missing docId")
		}
		return utils.HandleDeleteWithCtx(ctx, s3Client, bucketName)(c)
	})

	// Channel to listen for OS termination signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Println("Server started on :8080")
		if err := app.ListenTLS(":8080", "./certs/cert.pem", "./certs/key.pem"); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for termination signal
	<-quit
	log.Println("Graceful shutdown initiated...")

	// Stop accepting new connections, allow max 30s for active connections
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctxTimeout); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Cancel all active requests
	requestCtxs.Range(func(key, value any) bool {
		cancelFunc := value.(context.CancelFunc)
		cancelFunc()
		return true
	})

	// Wait for all active requests to finish
	for atomic.LoadInt32(&activeRequests) > 0 {
		log.Printf("Waiting for %d active requests to finish...", atomic.LoadInt32(&activeRequests))
		time.Sleep(500 * time.Millisecond)
	}

	log.Println("All active requests completed. Server exited gracefully.")
}
