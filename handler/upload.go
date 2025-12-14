package handler

import (
	"stream-download/service"

	"github.com/gofiber/fiber/v2"
)

type UploadHandler struct {
	s3Service *service.S3Service
}

func NewUploadHandler(s3Service *service.S3Service) *UploadHandler {
	return &UploadHandler{s3Service: s3Service}
}

type UploadResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    *UploadData `json:"data,omitempty"`
}

type UploadData struct {
	Filename string `json:"filename"`
	FilePath string `json:"filePath"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
}

// Upload godoc
// @Summary Upload file to S3
// @Description Upload a file to S3 bucket with optional custom path and filename
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param filePath formData string false "Custom path in S3 (e.g., 'images/2024')"
// @Param fileName formData string false "Custom filename (default: original filename)"
// @Success 200 {object} UploadResponse
// @Failure 400 {object} UploadResponse
// @Failure 500 {object} UploadResponse
// @Router /api/upload [post]
func (h *UploadHandler) Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(UploadResponse{
			Success: false,
			Message: "File is required",
		})
	}

	filePath := c.FormValue("filePath", "")
	fileName := c.FormValue("fileName", file.Filename)

	key := fileName
	if filePath != "" {
		key = filePath + "/" + fileName
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(UploadResponse{
			Success: false,
			Message: "Failed to open file",
		})
	}
	defer src.Close()

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	url, err := h.s3Service.Upload(c.Context(), key, src, contentType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(UploadResponse{
			Success: false,
			Message: "Failed to upload file: " + err.Error(),
		})
	}

	return c.JSON(UploadResponse{
		Success: true,
		Message: "File uploaded successfully",
		Data: &UploadData{
			Filename: fileName,
			FilePath: filePath,
			URL:      url,
			Size:     file.Size,
		},
	})
}
