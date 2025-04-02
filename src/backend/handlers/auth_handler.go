package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Eduardo-Barreto/web-ponderada/backend/auth"
	"github.com/Eduardo-Barreto/web-ponderada/backend/models"
	"github.com/Eduardo-Barreto/web-ponderada/backend/repository"
	"github.com/Eduardo-Barreto/web-ponderada/backend/utils"
)

type AuthHandler struct {
	UserRepo repository.UserRepository
}

func NewAuthHandler(userRepo repository.UserRepository) *AuthHandler {
	return &AuthHandler{UserRepo: userRepo}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var input models.UserRegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	// Check if user already exists
	existingUser, err := h.UserRepo.GetUserByEmail(context.Background(), input.Email)
	if err != nil {
		// Log the repository error but send a generic server error to the client
		log.Printf("Error checking existing user: %v", err)
		utils.SendError(c, http.StatusInternalServerError, "Failed to check user existence")
		return
	}
	if existingUser != nil {
		utils.SendError(c, http.StatusConflict, "User with this email already exists")
		return
	}

	// Hash the password
	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		utils.SendError(c, http.StatusInternalServerError, "Failed to process registration")
		return
	}

	// Create the user model
	newUser := &models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: hashedPassword, // Store the hash
		// ProfilePic is initially empty or default
	}

	// Save user to database
	userID, err := h.UserRepo.CreateUser(context.Background(), newUser)
	if err != nil {
		log.Printf("Error creating user in db: %v", err)
		utils.SendError(c, http.StatusInternalServerError, "Failed to register user")
		return
	}

	// Exclude password before sending response
	newUser.Password = ""
	newUser.ID = userID // Set the returned ID

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user": newUser})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	// Find user by email
	user, err := h.UserRepo.GetUserByEmail(context.Background(), input.Email)
	if err != nil {
		log.Printf("Error fetching user by email: %v", err)
		utils.SendError(c, http.StatusInternalServerError, "Error during login")
		return
	}
	if user == nil {
		utils.SendError(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Check password
	if !auth.CheckPasswordHash(input.Password, user.Password) {
		utils.SendError(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		log.Printf("Error generating JWT token: %v", err)
		utils.SendError(c, http.StatusInternalServerError, "Failed to login")
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
