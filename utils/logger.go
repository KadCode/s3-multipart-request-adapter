package utils

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// logRequest logs each incoming request with metadata and processing duration
func logRequest(c *fiber.Ctx, start time.Time, extra string) {
	requestID := c.Get("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	duration := time.Since(start).Round(time.Millisecond)
	log.Printf("[%s] req=%s %s %s func=%s ip=%s duration=%v %s",
		time.Now().Format(time.RFC3339),
		requestID,
		c.Method(),
		c.Path(),
		c.Query("funcName"),
		c.IP(),
		duration,
		extra,
	)
}
