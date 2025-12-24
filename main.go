package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"stream-download/config"
	_ "stream-download/docs"
	"stream-download/handler"
	"stream-download/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
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

	// Memory stats endpoint
	app.Get("/debug/memory", func(c *fiber.Ctx) error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return c.JSON(fiber.Map{
			"alloc_mb":       float64(m.Alloc) / 1024 / 1024,
			"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
			"sys_mb":         float64(m.Sys) / 1024 / 1024,
			"heap_alloc_mb":  float64(m.HeapAlloc) / 1024 / 1024,
			"heap_sys_mb":    float64(m.HeapSys) / 1024 / 1024,
			"heap_inuse_mb":  float64(m.HeapInuse) / 1024 / 1024,
			"num_gc":         m.NumGC,
			"goroutines":     runtime.NumGoroutine(),
		})
	})

	// pprof endpoints
	app.Get("/debug/pprof/*", adaptor.HTTPHandler(http.DefaultServeMux))

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
