package handlers

import (
	"log"
	"net/http"
	"os" // Use filepath for safety
	"strings"

	"github.com/Eduardo-Barreto/web-ponderada/backend/repository"
	"github.com/Eduardo-Barreto/web-ponderada/backend/utils"
	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	FileRepo repository.StorageRepository
}

func NewImageHandler(fileRepo repository.StorageRepository) *ImageHandler {
	return &ImageHandler{FileRepo: fileRepo}
}

// ServeImage serves static image files stored by the application
func (h *ImageHandler) ServeImage(c *gin.Context) {
	// Expecting path like /images/users/uuid.jpg or /images/products/uuid.png
	// We capture the subdirectory and filename together
	imgPath := c.Param("filepath") // e.g., "users/abc.jpg" or "products/xyz.png"
	log.Println("imgPath", imgPath)

	// Basic security: Prevent path traversal and absolute paths
	if strings.Contains(imgPath, "..") {
		utils.SendError(c, http.StatusBadRequest, "Invalid file path")
		return
	}

	// Use the storage repository to get the full system path
	fullPath := h.FileRepo.GetFilePath(imgPath)
	if fullPath == "" { // Check if GetFilePath considered it invalid
		utils.SendError(c, http.StatusBadRequest, "Invalid file path provided")
		return
	}

	// Check if the file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		log.Printf("Image not found: %s (requested path: %s)", fullPath, imgPath)
		utils.SendError(c, http.StatusNotFound, "Image not found")
		return
	}

	// Serve the file
	// Gin's c.File() handles setting Content-Type based on extension
	c.File(fullPath)
}
