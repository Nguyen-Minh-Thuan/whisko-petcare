package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"whisko-petcare/internal/infrastructure/cloudinary"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🚀 Cloudinary Example - Running...")
	fmt.Println("=" + string(make([]byte, 50)))
	fmt.Println()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  Warning: .env file not found: %v", err)
		fmt.Println()
	}

	// Example 1: Initialize Cloudinary service
	fmt.Println("📋 Example 1: Initialize Cloudinary Service")
	fmt.Println("-" + string(make([]byte, 50)))
	
	cloudinaryConfig, err := cloudinary.NewConfigFromEnv()
	if err != nil {
		fmt.Printf("❌ Failed to load Cloudinary config: %v\n", err)
		fmt.Println()
		fmt.Println("💡 To fix this:")
		fmt.Println("   1. Go to https://cloudinary.com/console")
		fmt.Println("   2. Copy your credentials")
		fmt.Println("   3. Add to .env file:")
		fmt.Println()
		fmt.Println("      CLOUDINARY_CLOUD_NAME=your-cloud-name")
		fmt.Println("      CLOUDINARY_API_KEY=your-api-key")
		fmt.Println("      CLOUDINARY_API_SECRET=your-api-secret")
		fmt.Println("      CLOUDINARY_UPLOAD_FOLDER=whisko-petcare")
		fmt.Println()
		os.Exit(1)
	}

	fmt.Printf("✅ Config loaded successfully\n")
	fmt.Printf("   Cloud Name: %s\n", cloudinaryConfig.CloudName)
	fmt.Printf("   API Key: %s...\n", maskString(cloudinaryConfig.APIKey, 6))
	fmt.Printf("   Upload Folder: %s\n", cloudinaryConfig.UploadFolder)
	fmt.Println()

	// Example 2: Create Cloudinary service
	fmt.Println("📋 Example 2: Create Cloudinary Service")
	fmt.Println("-" + string(make([]byte, 50)))
	
	cloudinaryService, err := cloudinary.NewService(cloudinaryConfig)
	if err != nil {
		fmt.Printf("❌ Failed to create Cloudinary service: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Cloudinary service created successfully")
	fmt.Println()

	// Example 3: Demonstrate URL generation
	fmt.Println("📋 Example 3: Generate Transformed URLs")
	fmt.Println("-" + string(make([]byte, 50)))
	
	// Using a sample public ID
	samplePublicID := "whisko-petcare/pets/sample_pet"
	
	ctx := context.Background()
	
	// Get thumbnail URL
	thumbnailURL, err := cloudinaryService.GetThumbnailURL(samplePublicID, 200)
	if err != nil {
		fmt.Printf("⚠️  Failed to generate thumbnail URL: %v\n", err)
	} else {
		fmt.Printf("✅ Thumbnail URL (200x200):\n   %s\n\n", thumbnailURL)
	}
	
	// Get optimized URL
	optimizedURL, err := cloudinaryService.GetOptimizedURL(samplePublicID, 800, 600)
	if err != nil {
		fmt.Printf("⚠️  Failed to generate optimized URL: %v\n", err)
	} else {
		fmt.Printf("✅ Optimized URL (800x600):\n   %s\n\n", optimizedURL)
	}
	
	// Get custom transformed URL
	customURL, err := cloudinaryService.GetTransformedURL(samplePublicID, "w_300,h_300,c_fill,g_face,r_max")
	if err != nil {
		fmt.Printf("⚠️  Failed to generate custom URL: %v\n", err)
	} else {
		fmt.Printf("✅ Custom transformed URL (circular crop with face focus):\n   %s\n\n", customURL)
	}

	// Example 4: Test public ID extraction
	fmt.Println("📋 Example 4: Extract Public ID from URL")
	fmt.Println("-" + string(make([]byte, 50)))
	
	sampleURL := "https://res.cloudinary.com/demo/image/upload/v1234567890/whisko-petcare/pets/pet_123.jpg"
	extractedID := cloudinaryService.ExtractPublicIDFromURL(sampleURL)
	fmt.Printf("Sample URL:\n   %s\n\n", sampleURL)
	fmt.Printf("✅ Extracted Public ID:\n   %s\n\n", extractedID)

	// Example 5: Demonstrate upload options structure
	fmt.Println("📋 Example 5: Upload Options Examples")
	fmt.Println("-" + string(make([]byte, 50)))
	
	// Pet upload options
	petOpts := &cloudinary.UploadOptions{
		Folder:         "pets",
		Tags:           []string{"pet", "profile"},
		ResourceType:   "image",
		AllowedFormats: []string{"jpg", "jpeg", "png", "webp"},
		Transformation: "q_auto,f_auto",
	}
	fmt.Printf("✅ Pet Upload Options:\n")
	fmt.Printf("   Folder: %s\n", petOpts.Folder)
	fmt.Printf("   Tags: %v\n", petOpts.Tags)
	fmt.Printf("   Allowed Formats: %v\n", petOpts.AllowedFormats)
	fmt.Printf("   Transformation: %s\n\n", petOpts.Transformation)

	// Vendor upload options
	vendorOpts := &cloudinary.UploadOptions{
		Folder:         "vendors",
		Tags:           []string{"vendor", "logo"},
		ResourceType:   "image",
		AllowedFormats: []string{"jpg", "jpeg", "png", "webp"},
		Transformation: "w_1200,h_800,c_fill,q_auto,f_auto",
	}
	fmt.Printf("✅ Vendor Upload Options:\n")
	fmt.Printf("   Folder: %s\n", vendorOpts.Folder)
	fmt.Printf("   Tags: %v\n", vendorOpts.Tags)
	fmt.Printf("   Transformation: %s\n\n", vendorOpts.Transformation)

	// Summary
	fmt.Println("=" + string(make([]byte, 50)))
	fmt.Println("✅ All examples completed successfully!")
	fmt.Println()
	fmt.Println("📚 Next Steps:")
	fmt.Println("   1. Test actual image upload with a real file")
	fmt.Println("   2. Integrate into your main.go application")
	fmt.Println("   3. Add image fields to your domain aggregates")
	fmt.Println("   4. Update command handlers to use Cloudinary")
	fmt.Println()
	fmt.Println("📖 Documentation:")
	fmt.Println("   - Quick Reference: docs/CLOUDINARY_QUICK_REFERENCE.md")
	fmt.Println("   - Full Guide: docs/CLOUDINARY_INTEGRATION.md")
	fmt.Println("   - Setup Help: internal/infrastructure/cloudinary/README.md")
	fmt.Println()

	// Note about actual uploads
	fmt.Println("💡 Note: This example doesn't upload actual files to avoid")
	fmt.Println("   consuming your Cloudinary quota. To test real uploads:")
	fmt.Println()
	fmt.Println("   1. Start your API server: go run cmd/api/main.go")
	fmt.Println("   2. Use curl to upload:")
	fmt.Println()
	fmt.Println("      curl -X POST http://localhost:8080/api/images/upload \\")
	fmt.Println("        -F \"image=@/path/to/image.jpg\" \\")
	fmt.Println("        -F \"entity_type=pet\" \\")
	fmt.Println("        -F \"entity_id=test123\"")
	fmt.Println()

	_ = ctx // avoid unused variable warning
}

func maskString(s string, visibleChars int) string {
	if len(s) <= visibleChars {
		return "***"
	}
	if len(s) <= visibleChars+4 {
		return s[:visibleChars] + "***"
	}
	return s[:visibleChars] + "..." + s[len(s)-3:]
}
