package main

import (
	"log"

	"stream-download/config"
	_ "stream-download/docs"
	"stream-download/handler"
	"stream-download/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// @title S3 File Upload API
// @version 1.0
// @description API untuk upload file ke S3 menggunakan LocalStack
// @host localhost:8080
// @BasePath /

func main() {
	cfg := config.Load()

	s3Service, err := service.NewS3Service(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize S3 service: %v", err)
	}

	uploadHandler := handler.NewUploadHandler(s3Service)

	app := fiber.New(fiber.Config{
		BodyLimit: cfg.MaxUploadSize * 1024 * 1024, // MB to bytes
	})

	app.Get("/swagger/*", swagger.HandlerDefault)
	app.Static("/", "./static")

	api := app.Group("/api")
	api.Post("/upload", uploadHandler.Upload)
	api.Post("/download", uploadHandler.Download)

	log.Printf("Server running on http://localhost:%s", cfg.ServerPort)
	log.Printf("Swagger UI: http://localhost:%s/swagger/", cfg.ServerPort)
	log.Printf("Streaming Test: http://localhost:%s/", cfg.ServerPort)

	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
