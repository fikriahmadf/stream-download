package handler

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type DownloadRequest struct {
	URLs []string `json:"urls"`
}

// Download godoc
// @Summary Download files from URLs and zip them
// @Description Download multiple files from URLs and return as a zip file
// @Tags download
// @Accept json
// @Produce application/zip
// @Param request body DownloadRequest true "List of file URLs to download"
// @Success 200 {file} binary "Zip file containing all requested files"
// @Failure 400 {object} UploadResponse
// @Failure 500 {object} UploadResponse
// @Router /api/download [post]
func (h *UploadHandler) Download(c *fiber.Ctx) error {
	var req DownloadRequest

	// Support both JSON body and form data
	contentType := c.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(UploadResponse{
				Success: false,
				Message: "Invalid request body: " + err.Error(),
			})
		}
	} else {
		// Form submit - parse JSON from form field
		jsonData := c.FormValue("json")
		if jsonData != "" {
			if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(UploadResponse{
					Success: false,
					Message: "Invalid JSON in form: " + err.Error(),
				})
			}
		}
	}

	if len(req.URLs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(UploadResponse{
			Success: false,
			Message: "URLs array is required",
		})
	}

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", "attachment; filename=download.zip")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		zipWriter := zip.NewWriter(w)
		defer zipWriter.Close()

		for _, fileURL := range req.URLs {
			resp, err := http.Get(fileURL)
			if err != nil {
				continue
			}

			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				continue
			}

			fileName := path.Base(fileURL)
			zipFile, err := zipWriter.Create(fileName)
			if err != nil {
				resp.Body.Close()
				continue
			}

			io.Copy(zipFile, resp.Body)
			resp.Body.Close()

			zipWriter.Flush()
			w.Flush()
		}
	})

	return nil
}
