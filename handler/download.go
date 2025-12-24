package handler

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

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

		var failedFiles []string
		successCount := 0

		for _, fileURL := range req.URLs {
			fileName := path.Base(fileURL)

			resp, err := http.Get(fileURL)
			if err != nil {
				failedFiles = append(failedFiles, fmt.Sprintf("%s: %s", fileName, err.Error()))
				continue
			}

			if resp.StatusCode != http.StatusOK {
				failedFiles = append(failedFiles, fmt.Sprintf("%s: HTTP %d", fileName, resp.StatusCode))
				resp.Body.Close()
				continue
			}

			zipFile, err := zipWriter.Create(fileName)
			if err != nil {
				failedFiles = append(failedFiles, fmt.Sprintf("%s: failed to create zip entry - %s", fileName, err.Error()))
				resp.Body.Close()
				continue
			}

			_, err = io.Copy(zipFile, resp.Body)
			resp.Body.Close()

			if err != nil {
				failedFiles = append(failedFiles, fmt.Sprintf("%s: failed to write - %s", fileName, err.Error()))
				continue
			}

			successCount++
			zipWriter.Flush()
			w.Flush()
		}

		// If there are failed files, add an error report to the ZIP
		if len(failedFiles) > 0 {
			errorFile, err := zipWriter.Create("_download_errors.txt")
			if err == nil {
				var report strings.Builder
				report.WriteString("Download Error Report\n")
				report.WriteString("=====================\n")
				report.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))
				report.WriteString(fmt.Sprintf("Total requested: %d\n", len(req.URLs)))
				report.WriteString(fmt.Sprintf("Successful: %d\n", successCount))
				report.WriteString(fmt.Sprintf("Failed: %d\n\n", len(failedFiles)))
				report.WriteString("Failed files:\n")
				for _, f := range failedFiles {
					report.WriteString(fmt.Sprintf("  - %s\n", f))
				}
				errorFile.Write([]byte(report.String()))
			}
		}

		// If ALL files failed, write a prominent error file
		if successCount == 0 && len(req.URLs) > 0 {
			allFailedFile, err := zipWriter.Create("_ALL_DOWNLOADS_FAILED.txt")
			if err == nil {
				allFailedFile.Write([]byte("ERROR: All file downloads failed. Please check _download_errors.txt for details.\n"))
			}
		}
	})

	return nil
}
