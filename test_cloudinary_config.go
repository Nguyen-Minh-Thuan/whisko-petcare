package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Quick test to verify Cloudinary credentials are properly configured
func main() {
	fmt.Println("🔍 Cloudinary Configuration Test")
	fmt.Println("==================================")
	fmt.Println()

	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("⚠️  Warning: .env file not found or could not be loaded\n")
		fmt.Printf("   Error: %v\n\n", err)
	} else {
		fmt.Println("✅ .env file loaded successfully")
		fmt.Println()
	}

	// Check required environment variables
	fmt.Println("📋 Checking Cloudinary Environment Variables:")
	fmt.Println()

	required := map[string]string{
		"CLOUDINARY_CLOUD_NAME": "Your Cloudinary cloud name",
		"CLOUDINARY_API_KEY":    "Your Cloudinary API key",
		"CLOUDINARY_API_SECRET": "Your Cloudinary API secret",
	}

	optional := map[string]string{
		"CLOUDINARY_UPLOAD_FOLDER": "Upload folder (default: whisko-petcare)",
	}

	allSet := true

	// Check required variables
	for key, description := range required {
		value := os.Getenv(key)
		if value == "" {
			fmt.Printf("❌ %s: NOT SET\n", key)
			fmt.Printf("   → %s\n", description)
			allSet = false
		} else {
			// Mask sensitive values
			maskedValue := maskValue(value, key)
			fmt.Printf("✅ %s: %s\n", key, maskedValue)
		}
		fmt.Println()
	}

	// Check optional variables
	fmt.Println("Optional Variables:")
	for key, description := range optional {
		value := os.Getenv(key)
		if value == "" {
			fmt.Printf("ℹ️  %s: NOT SET (will use default)\n", key)
			fmt.Printf("   → %s\n", description)
		} else {
			fmt.Printf("✅ %s: %s\n", key, value)
		}
		fmt.Println()
	}

	// Final verdict
	fmt.Println("==================================")
	if allSet {
		fmt.Println("✅ All required Cloudinary credentials are configured!")
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("1. Run your application: go run cmd/api/main.go")
		fmt.Println("2. Test image upload: See docs/CLOUDINARY_QUICK_REFERENCE.md")
		fmt.Println("3. Check logs for: '✅ Cloudinary service initialized'")
	} else {
		fmt.Println("❌ Some required credentials are missing!")
		fmt.Println()
		fmt.Println("How to fix:")
		fmt.Println("1. Go to https://cloudinary.com/console")
		fmt.Println("2. Copy your Cloud Name, API Key, and API Secret")
		fmt.Println("3. Add them to your .env file:")
		fmt.Println()
		fmt.Println("   CLOUDINARY_CLOUD_NAME=your-cloud-name")
		fmt.Println("   CLOUDINARY_API_KEY=your-api-key")
		fmt.Println("   CLOUDINARY_API_SECRET=your-api-secret")
		fmt.Println("   CLOUDINARY_UPLOAD_FOLDER=whisko-petcare")
		fmt.Println()
		fmt.Println("4. Run this test again: go run test_cloudinary_config.go")
	}
	fmt.Println("==================================")
}

func maskValue(value, key string) string {
	if key == "CLOUDINARY_API_SECRET" {
		if len(value) > 8 {
			return value[:4] + "..." + value[len(value)-4:]
		}
		return "***"
	}
	if key == "CLOUDINARY_API_KEY" {
		if len(value) > 8 {
			return value[:6] + "..." + value[len(value)-4:]
		}
		return "***"
	}
	return value
}
