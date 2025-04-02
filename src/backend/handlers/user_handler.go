package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Eduardo-Barreto/web-ponderada/backend/models"
	"github.com/Eduardo-Barreto/web-ponderada/backend/repository"
	"github.com/Eduardo-Barreto/web-ponderada/backend/utils"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	UserRepo repository.UserRepository
	FileRepo repository.StorageRepository // Inject file repo for profile pics
}

func NewUserHandler(userRepo repository.UserRepository, fileRepo repository.StorageRepository) *UserHandler {
	return &UserHandler{UserRepo: userRepo, FileRepo: fileRepo}
}

// GetUsers retrieves a list of all users
func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.UserRepo.GetAllUsers(context.Background())
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		utils.SendError(c, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetUser retrieves a single user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	user, err := h.UserRepo.GetUserByID(context.Background(), id)
	if err != nil {
		// Check if the error is "user not found"
		if err.Error() == "user not found" {
			utils.SendError(c, http.StatusNotFound, err.Error())
		} else {
			log.Printf("Error getting user by ID %d: %v", id, err)
			utils.SendError(c, http.StatusInternalServerError, "Failed to retrieve user")
		}
		return
	}

	// Don't send password hash
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

// UpdateUser handles updating user information (name, email)
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	// --- Authorization Check ---
	// Get user ID from token (set by middleware)
	tokenUserID, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "Unauthorized access")
		return
	}
	// Check if the user making the request is the user being updated
	if tokenUserID.(int) != id {
		utils.SendError(c, http.StatusForbidden, "You can only update your own profile")
		return
	}
	// --- End Authorization Check ---

	var input models.UserUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid input data: "+err.Error())
		return
	}

	// Prevent updating fields that aren't allowed or empty request
	if input.Name == nil && input.Email == nil {
		utils.SendError(c, http.StatusBadRequest, "No update fields provided")
		return
	}

	err = h.UserRepo.UpdateUser(context.Background(), id, &input)
	if err != nil {
		if err.Error() == "user not found or no changes made" {
			utils.SendError(c, http.StatusNotFound, "User not found or no changes were necessary")
		} else if strings.Contains(err.Error(), "unique constraint") { // Basic check for unique email error
			utils.SendError(c, http.StatusConflict, "Email address already in use")
		} else {
			log.Printf("Error updating user ID %d: %v", id, err)
			utils.SendError(c, http.StatusInternalServerError, "Failed to update user")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// UploadProfilePic handles uploading a new profile picture for a user
func (h *UserHandler) UploadProfilePic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	// --- Authorization Check ---
	tokenUserID, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "Unauthorized access")
		return
	}
	if tokenUserID.(int) != id {
		utils.SendError(c, http.StatusForbidden, "You can only update your own profile picture")
		return
	}
	// --- End Authorization Check ---

	file, err := c.FormFile("profile_pic") // Name of the form field
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Profile picture file is required: "+err.Error())
		return
	}

	// Optional: Delete old picture before saving new one
	currentUser, err := h.UserRepo.GetUserByID(context.Background(), id)
	if err == nil && currentUser.ProfilePic != "" {
		err := h.FileRepo.DeleteFile(context.Background(), currentUser.ProfilePic)
		if err != nil {
			// Log the error but don't necessarily block the upload
			log.Printf("Warning: Failed to delete old profile picture '%s' for user %d: %v", currentUser.ProfilePic, id, err)
		}
	} else if err != nil {
		log.Printf("Warning: Could not fetch user %d to check for old profile pic: %v", id, err)
	}

	// Save the file using the storage repository (e.g., to "uploads/users/")
	filename, err := h.FileRepo.SaveFile(context.Background(), file, "users")
	if err != nil {
		log.Printf("Error saving profile picture for user %d: %v", id, err)
		utils.SendError(c, http.StatusInternalServerError, "Failed to save profile picture: "+err.Error())
		return
	}

	// Update the user record in the database with the new filename
	err = h.UserRepo.UpdateUserProfilePic(context.Background(), id, filename)
	if err != nil {
		// If DB update fails, try to delete the just uploaded file
		h.FileRepo.DeleteFile(context.Background(), filename) // Best effort deletion
		log.Printf("Error updating user profile pic in DB for user %d: %v", id, err)
		if err.Error() == "user not found" {
			utils.SendError(c, http.StatusNotFound, "User not found")
		} else {
			utils.SendError(c, http.StatusInternalServerError, "Failed to update profile picture record")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile picture updated successfully", "filename": filename})
}

// DeleteUser handles deleting a user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	// --- Authorization Check ---
	// Maybe only admins can delete, or users can delete themselves
	tokenUserID, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "Unauthorized access")
		return
	}
	// For simplicity, allow users to delete themselves. Add admin logic if needed.
	if tokenUserID.(int) != id {
		utils.SendError(c, http.StatusForbidden, "You can only delete your own account")
		return
	}
	// --- End Authorization Check ---

	// Important: Delete associated profile picture file first!
	user, err := h.UserRepo.GetUserByID(context.Background(), id)
	if err != nil {
		if err.Error() == "user not found" {
			utils.SendError(c, http.StatusNotFound, err.Error())
		} else {
			log.Printf("Error finding user %d before deletion: %v", id, err)
			utils.SendError(c, http.StatusInternalServerError, "Failed to retrieve user before deletion")
		}
		return
	}

	if user.ProfilePic != "" {
		err := h.FileRepo.DeleteFile(context.Background(), user.ProfilePic)
		if err != nil {
			// Log error, but proceed with DB deletion anyway
			log.Printf("Warning: Failed to delete profile picture '%s' for user %d during deletion: %v", user.ProfilePic, id, err)
		}
	}

	err = h.UserRepo.DeleteUser(context.Background(), id)
	if err != nil {
		if err.Error() == "user not found" {
			utils.SendError(c, http.StatusNotFound, err.Error())
		} else {
			log.Printf("Error deleting user ID %d: %v", id, err)
			utils.SendError(c, http.StatusInternalServerError, "Failed to delete user")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
