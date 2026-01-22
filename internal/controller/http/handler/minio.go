package handler

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

type File struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type UploadedFile struct {
	FileName  string    `json:"file_name"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

// Img upload
// @Security    BearerAuth
// @Summary File upload
// @Description File upload
// @Tags Img-upload
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "File"
// @Router /img-upload [post]
// @Success 200 {object} string
func (h *Handler) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		slog.Error("Error uploading file", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File not provided"})
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	allowedExt := map[string]bool{
		".png":  true,
		".jpg":  true,
	}

	if !allowedExt[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Fayl turi qo'llab-quvvatlanmaydi: %s", ext)})
		return
	}

	id := uuid.NewString()
	fileName := id + header.Filename

	tempDir := os.TempDir()
	tempFilePath := filepath.Join(tempDir, fileName)

	out, err := os.Create(tempFilePath)
	if err != nil {
		slog.Error("Error creating temporary file", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create temporary file"})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		slog.Error("Error saving file", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	minioURL, err := h.MinIO.Upload(*h.Config, fileName, tempFilePath)
	if err != nil {
		slog.Error("Error uploading to MinIO", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to MinIO"})
		return
	}

	os.Remove(tempFilePath)

	c.JSON(http.StatusOK, gin.H{
		"Message":  "Successfully uploaded",
		"FileName": minioURL,
	})
}
