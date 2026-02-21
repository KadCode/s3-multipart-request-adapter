package utils

import (
	"log"

	s3_adapter_config "example.com/s3-multipart-request-adapter/config"
	"github.com/gofiber/fiber/v2"
)

func CreateNewFiberAppInstance() *fiber.App {
	cfg, err := s3_adapter_config.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return fiber.New(fiber.Config{
		Prefork:       cfg.FiberConfig.Prefork,
		CaseSensitive: cfg.FiberConfig.CaseSensitive,
		StrictRouting: cfg.FiberConfig.StrictRouting,
		ServerHeader:  cfg.FiberConfig.ServerHeader,
		AppName:       cfg.FiberConfig.AppName,
		ReadTimeout:   cfg.FiberConfig.ReadTimeout,
		BodyLimit:     cfg.FiberConfig.BodyLimit,
	})
}
