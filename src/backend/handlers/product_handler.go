package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Eduardo-Barreto/web-ponderada/backend/models"
	"github.com/Eduardo-Barreto/web-ponderada/backend/repository"
	"github.com/Eduardo-Barreto/web-ponderada/backend/utils"
)

type ProductHandler struct {
	ProductRepo repository.ProductRepository
	FileRepo    repository.StorageRepository
}

func NewProductHandler(productRepo repository.ProductRepository, fileRepo repository.StorageRepository) *ProductHandler {
	return &ProductHandler{ProductRepo: productRepo, FileRepo: fileRepo}
}

// CreateProduct handles creation of a new product, including image upload
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	// Use ShouldBind for form data, NOT ShouldBindJSON
	var input models.ProductInput
	// We bind form fields first (description, value, quantity)
	// Note: Binding multipart/form-data directly to struct with binding tags might be tricky.
	// It's often easier to parse fields manually.

	desc := c.PostForm("description")
	valueStr := c.PostForm("value")
	quantityStr := c.PostForm("quantity")

	// Manual Validation (or use a validation library compatible with form data)
	if desc == "" {
		utils.SendError(c, http.StatusBadRequest, "Description is required")
		return
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil || value <= 0 {
		utils.SendError(c, http.StatusBadRequest, "Invalid or non-positive value")
		return
	}
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil || quantity < 0 {
		utils.SendError(c, http.StatusBadRequest, "Invalid or negative quantity")
		return
	}

	input.Description = desc
	input.Value = value
	input.Quantity = quantity


	// Handle file upload (optional image)
	file, err := c.FormFile("image") // Name of the image field in the form
	var imageFilename string = "" // Default to empty string if no file

	if err == nil { // If a file was actually uploaded
		savedFilename, err := h.FileRepo.SaveFile(context.Background(), file, "products")
		if err != nil {
			log.Printf("Error saving product image: %v", err)
			utils.SendError(c, http.StatusInternalServerError, "Failed to save product image: "+err.Error())
			return
		}
		imageFilename = savedFilename
	} else if err != http.ErrMissingFile {
		// If there's an error other than the file simply not being there
		utils.SendError(c, http.StatusBadRequest, "Error processing image file: "+err.Error())
		return
	}

	// Create Product model
	newProduct := &models.Product{
		Description: input.Description,
		Value:       input.Value,
		Quantity:    input.Quantity,
		Image:       imageFilename, // Store the generated filename
	}

	// Save product to database
	productID, err := h.ProductRepo.CreateProduct(context.Background(), newProduct)
	if err != nil {
		log.Printf("Error creating product in db: %v", err)
        // If DB fails, delete the potentially uploaded file
        if imageFilename != "" {
            h.FileRepo.DeleteFile(context.Background(), imageFilename) // Best effort
        }
		utils.SendError(c, http.StatusInternalServerError, "Failed to create product")
		return
	}
	newProduct.ID = productID // Set the returned ID

	c.JSON(http.StatusCreated, newProduct)
}

// GetProducts retrieves a list of all products
func (h *ProductHandler) GetProducts(c *gin.Context) {
	products, err := h.ProductRepo.GetAllProducts(context.Background())
	if err != nil {
		log.Printf("Error getting all products: %v", err)
		utils.SendError(c, http.StatusInternalServerError, "Failed to retrieve products")
		return
	}
	c.JSON(http.StatusOK, products)
}

// GetProduct retrieves a single product by ID
func (h *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid product ID format")
		return
	}

	product, err := h.ProductRepo.GetProductByID(context.Background(), id)
	if err != nil {
		if err.Error() == "product not found" {
			utils.SendError(c, http.StatusNotFound, err.Error())
		} else {
			log.Printf("Error getting product by ID %d: %v", id, err)
			utils.SendError(c, http.StatusInternalServerError, "Failed to retrieve product")
		}
		return
	}
	c.JSON(http.StatusOK, product)
}

// UpdateProduct handles updating product details and potentially the image
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid product ID format")
		return
	}

	// --- Authorization: Ensure user is allowed (e.g., admin) - Skipped for brevity ---

	// Similar to Create, parse form fields manually
	desc := c.PostForm("description")
	valueStr := c.PostForm("value")
	quantityStr := c.PostForm("quantity")

	// Validation
	if desc == "" {
		utils.SendError(c, http.StatusBadRequest, "Description is required")
		return
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil || value <= 0 {
		utils.SendError(c, http.StatusBadRequest, "Invalid or non-positive value")
		return
	}
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil || quantity < 0 {
		utils.SendError(c, http.StatusBadRequest, "Invalid or negative quantity")
		return
	}

	input := models.ProductInput{
		Description: desc,
		Value:       value,
		Quantity:    quantity,
	}

    // Check for existing product to potentially delete old image
    oldProduct, err := h.ProductRepo.GetProductByID(context.Background(), id)
    if err != nil {
         if err.Error() == "product not found" {
			utils.SendError(c, http.StatusNotFound, err.Error())
		} else {
			log.Printf("Error getting product %d before update: %v", id, err)
			utils.SendError(c, http.StatusInternalServerError, "Failed to retrieve product before update")
		}
        return
    }

	// Handle optional new image upload
	file, err := c.FormFile("image")
	var newImageFilename *string // Pointer to store new filename if uploaded

	if err == nil { // New image provided
		savedFilename, err := h.FileRepo.SaveFile(context.Background(), file, "products")
		if err != nil {
			log.Printf("Error saving updated product image: %v", err)
			utils.SendError(c, http.StatusInternalServerError, "Failed to save updated image: "+err.Error())
			return
		}
		newImageFilename = &savedFilename // Store pointer to new filename

        // Delete the old image *after* successfully saving the new one
        if oldProduct.Image != "" {
            errDelete := h.FileRepo.DeleteFile(context.Background(), oldProduct.Image)
             if errDelete != nil {
                 log.Printf("Warning: Failed to delete old image '%s' during product %d update: %v", oldProduct.Image, id, errDelete)
                 // Proceed with DB update anyway
            }
        }

	} else if err != http.ErrMissingFile {
		utils.SendError(c, http.StatusBadRequest, "Error processing image file: "+err.Error())
		return
	}
    // If err == http.ErrMissingFile, newImageFilename remains nil, so DB won't update the image field


	// Update product in database
	err = h.ProductRepo.UpdateProduct(context.Background(), id, &input, newImageFilename)
	if err != nil {
        // If DB update fails, delete the *newly* uploaded file (if any)
        if newImageFilename != nil {
            h.FileRepo.DeleteFile(context.Background(), *newImageFilename) // Best effort
        }

		if err.Error() == "product not found or no changes made" {
			utils.SendError(c, http.StatusNotFound, "Product not found or no changes necessary")
		} else {
			log.Printf("Error updating product ID %d: %v", id, err)
			utils.SendError(c, http.StatusInternalServerError, "Failed to update product")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}

// DeleteProduct handles deleting a product
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid product ID format")
		return
	}

	// --- Authorization Check (e.g., Admin only) - Skipped ---

    // Get product details to delete the image first
    product, err := h.ProductRepo.GetProductByID(context.Background(), id)
    if err != nil {
         if err.Error() == "product not found" {
			utils.SendError(c, http.StatusNotFound, err.Error())
		} else {
			log.Printf("Error getting product %d before deletion: %v", id, err)
			utils.SendError(c, http.StatusInternalServerError, "Failed to retrieve product before deletion")
		}
        return
    }


    // Delete the associated image file from storage
    if product.Image != "" {
        err := h.FileRepo.DeleteFile(context.Background(), product.Image)
        if err != nil {
            // Log error but proceed with DB deletion
             log.Printf("Warning: Failed to delete image file '%s' for product %d during deletion: %v", product.Image, id, err)
        }
    }


	// Delete product from database
	err = h.ProductRepo.DeleteProduct(context.Background(), id)
	if err != nil {
		if err.Error() == "product not found" {
			utils.SendError(c, http.StatusNotFound, err.Error())
		} else {
			log.Printf("Error deleting product ID %d: %v", id, err)
			utils.SendError(c, http.StatusInternalServerError, "Failed to delete product")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
