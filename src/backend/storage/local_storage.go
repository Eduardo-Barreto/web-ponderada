package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/Eduardo-Barreto/web-ponderada/backend/config"
	"github.com/Eduardo-Barreto/web-ponderada/backend/repository"
)

type localStorage struct {
	uploadDir string
}

// NewLocalStorage creates a repository for local file storage
func NewLocalStorage() repository.StorageRepository {
	return &localStorage{
		uploadDir: config.AppConfig.UploadDir,
	}
}

// SaveFile saves the uploaded file to the local disk and returns the generated filename
func (s *localStorage) SaveFile(ctx context.Context, fileHeader *multipart.FileHeader, destinationSubDir string) (string, error) {
	// Generate a unique filename to prevent collisions and hide original name
	ext := filepath.Ext(fileHeader.Filename)
	// Basic validation for allowed extensions (example)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	if !allowedExts[strings.ToLower(ext)] {
		return "", fmt.Errorf("invalid file type: %s", ext)
	}

	newFileName := uuid.New().String() + ext

	// Create subdirectory if it doesn't exist
	fullDirPath := filepath.Join(s.uploadDir, destinationSubDir)
	if _, err := os.Stat(fullDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(fullDirPath, 0755); err != nil {
			log.Printf("Error creating subdirectory %s: %v", fullDirPath, err)
			return "", fmt.Errorf("could not create storage directory")
		}
	}


	// Construct the full path including the potential subdirectory
	// The filename stored in DB should include the subdirectory if used
	relativePath := filepath.Join(destinationSubDir, newFileName) // e.g., "users/uuid.jpg" or "products/uuid.png"
	fullPath := filepath.Join(s.uploadDir, relativePath)

	// Open the uploaded file
	src, err := fileHeader.Open()
	if err != nil {
		log.Printf("Error opening uploaded file: %v", err)
		return "", fmt.Errorf("failed to open uploaded file")
	}
	defer src.Close()

	// Create the destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		log.Printf("Error creating destination file %s: %v", fullPath, err)
		return "", fmt.Errorf("failed to save file")
	}
	defer dst.Close()

	// Copy the file content
	if _, err = io.Copy(dst, src); err != nil {
		log.Printf("Error copying file content to %s: %v", fullPath, err)
		// Attempt to remove partially written file
		os.Remove(fullPath)
		return "", fmt.Errorf("failed to copy file content")
	}

	log.Printf("Successfully saved file: %s", fullPath)
	// Return the relative path (including subdirectory) to be stored in DB
	return relativePath, nil
}

// GetFilePath returns the absolute path on the server for a given filename
func (s *localStorage) GetFilePath(filename string) string {
    // Basic check to prevent path traversal
    cleanFilename := filepath.Clean(filename)
    if strings.Contains(cleanFilename, "..") {
        log.Printf("Warning: Potential path traversal attempt detected: %s", filename)
        return "" // Return empty string or handle as an error
    }
	return filepath.Join(s.uploadDir, cleanFilename)
}

// DeleteFile removes a file from the local storage
func (s *localStorage) DeleteFile(ctx context.Context, filename string) error {
	filePath := s.GetFilePath(filename)
    if filePath == "" { // Check result from GetFilePath
        return fmt.Errorf("invalid filename provided: %s", filename)
    }

	// Check if file exists before attempting deletion
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("File not found for deletion, possibly already deleted: %s", filePath)
		return nil // Not an error if it doesn't exist
	}

	err := os.Remove(filePath)
	if err != nil {
		log.Printf("Error deleting file %s: %v", filePath, err)
		return fmt.Errorf("failed to delete file")
	}
	log.Printf("Successfully deleted file: %s", filePath)
	return nil
}
