package examples

import (
	"context"
	"log"
	"net/http"

	"whisko-petcare/internal/infrastructure/cloudinary"
)

// Example of how to integrate Cloudinary into your main.go application

func integrateCloudinaryExample() {
	// 1. Initialize Cloudinary service from environment variables
	cloudinaryConfig, err := cloudinary.NewConfigFromEnv()
	if err != nil {
		log.Printf("Warning: Cloudinary not configured: %v", err)
		log.Println("To enable Cloudinary, set these environment variables:")
		log.Println("  CLOUDINARY_CLOUD_NAME")
		log.Println("  CLOUDINARY_API_KEY")
		log.Println("  CLOUDINARY_API_SECRET")
		return
	}

	cloudinaryService, err := cloudinary.NewService(cloudinaryConfig)
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary service: %v", err)
	}
	log.Println("✅ Cloudinary service initialized")

	// 2. Initialize Cloudinary HTTP handler
	cloudinaryHandler := cloudinary.NewHandler(cloudinaryService)

	// 3. Register routes (adjust based on your router)
	// Using standard http.ServeMux as example:
	mux := http.NewServeMux()
	
	mux.HandleFunc("/api/images/upload", cloudinaryHandler.HandleUploadImage)
	mux.HandleFunc("/api/images/delete", cloudinaryHandler.HandleDeleteImage)
	mux.HandleFunc("/api/images/transform", cloudinaryHandler.HandleGetTransformedURL)

	log.Println("Cloudinary routes registered:")
	log.Println("  POST   /api/images/upload")
	log.Println("  DELETE /api/images/delete")
	log.Println("  POST   /api/images/transform")
}

// Example of using Cloudinary service in command handlers
func exampleUsageInCommandHandler(cloudinaryService *cloudinary.Service) {
	ctx := context.Background()

	// Example 1: Upload pet image from multipart form
	// (Assuming you have file, filename from request)
	/*
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		return err
	}
	defer file.Close()

	result, err := cloudinaryService.UploadPetImage(ctx, file, fileHeader.Filename, petID)
	if err != nil {
		return fmt.Errorf("failed to upload pet image: %w", err)
	}

	// Save the image URL and public ID to your pet aggregate
	pet.ImageURL = result.SecureURL
	pet.ImagePublicID = result.PublicID
	*/

	// Example 2: Delete old image when updating
	/*
	if pet.ImagePublicID != "" {
		if err := cloudinaryService.DeleteFile(ctx, pet.ImagePublicID); err != nil {
			log.Printf("Warning: failed to delete old image: %v", err)
		}
	}
	*/

	// Example 3: Generate thumbnail URL
	/*
	thumbnailURL, err := cloudinaryService.GetThumbnailURL(pet.ImagePublicID, 200)
	if err != nil {
		log.Printf("Failed to generate thumbnail: %v", err)
	}
	*/

	// Example 4: Custom upload with specific options
	/*
	opts := &cloudinary.UploadOptions{
		Folder:         "services",
		Tags:           []string{"service", serviceID},
		UniqueFilename: true,
		ResourceType:   "image",
		AllowedFormats: []string{"jpg", "jpeg", "png", "webp"},
		Transformation: "w_1200,h_800,c_fill,q_auto,f_auto",
	}
	
	result, err := cloudinaryService.UploadFile(ctx, file, filename, opts)
	if err != nil {
		return fmt.Errorf("failed to upload service image: %w", err)
	}
	*/

	log.Println("See function body for usage examples")
	_ = ctx
}

// Example: Add Cloudinary service to your application command handlers
type PetCommandHandlerWithCloudinary struct {
	// ... existing fields
	cloudinaryService *cloudinary.Service
}

// Example: Update pet with image upload
func (h *PetCommandHandlerWithCloudinary) HandleUpdatePetWithImage(ctx context.Context, petID string, imageFile interface{}) error {
	// 1. Upload the new image
	/*
	newResult, err := h.cloudinaryService.UploadPetImage(ctx, imageFile, filename, petID)
	if err != nil {
		return fmt.Errorf("failed to upload new image: %w", err)
	}

	// 2. Get the existing pet to retrieve old image
	pet, err := h.petRepository.GetByID(ctx, petID)
	if err != nil {
		// Clean up the newly uploaded image since we failed
		h.cloudinaryService.DeleteFile(ctx, newResult.PublicID)
		return fmt.Errorf("failed to get pet: %w", err)
	}

	// 3. Delete the old image if it exists
	if pet.ImagePublicID != "" {
		if err := h.cloudinaryService.DeleteFile(ctx, pet.ImagePublicID); err != nil {
			log.Printf("Warning: failed to delete old image %s: %v", pet.ImagePublicID, err)
		}
	}

	// 4. Update pet with new image URL
	pet.ImageURL = newResult.SecureURL
	pet.ImagePublicID = newResult.PublicID

	// 5. Save the updated pet
	if err := h.petRepository.Update(ctx, pet); err != nil {
		// Try to clean up the new image
		h.cloudinaryService.DeleteFile(ctx, newResult.PublicID)
		return fmt.Errorf("failed to update pet: %w", err)
	}

	return nil
	*/

	return nil
}

// Integration checklist:
// 
// 1. ✅ Add Cloudinary credentials to .env file
// 2. ✅ Initialize Cloudinary service in main.go
// 3. ✅ Register Cloudinary HTTP routes
// 4. ✅ Add ImageURL and ImagePublicID fields to domain aggregates (Pet, Vendor, User, Service)
// 5. ✅ Update command handlers to use Cloudinary service
// 6. ✅ Update projections to include image fields
// 7. ✅ Add cleanup logic to delete images when entities are deleted
// 8. ✅ Test image upload/delete/transform functionality
// 9. ✅ Update frontend to handle image uploads
// 10. ✅ Implement error handling and rollback for failed uploads
