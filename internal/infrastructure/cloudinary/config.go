package cloudinary

import (
	"fmt"
	"os"
)

// Config holds Cloudinary configuration
type Config struct {
	CloudName string
	APIKey    string
	APISecret string
	UploadFolder string // Optional: default folder for uploads
}

// NewConfigFromEnv creates a new Cloudinary config from environment variables
func NewConfigFromEnv() (*Config, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("missing required Cloudinary environment variables (CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY, CLOUDINARY_API_SECRET)")
	}

	uploadFolder := os.Getenv("CLOUDINARY_UPLOAD_FOLDER")
	if uploadFolder == "" {
		// Fallback to CLOUDINARY_FOLDER for compatibility
		uploadFolder = os.Getenv("CLOUDINARY_FOLDER")
	}
	if uploadFolder == "" {
		uploadFolder = "whisko-petcare" // default folder
	}

	return &Config{
		CloudName:    cloudName,
		APIKey:       apiKey,
		APISecret:    apiSecret,
		UploadFolder: uploadFolder,
	}, nil
}

// Validate checks if the config is valid
func (c *Config) Validate() error {
	if c.CloudName == "" {
		return fmt.Errorf("cloudinary cloud name is required")
	}
	if c.APIKey == "" {
		return fmt.Errorf("cloudinary API key is required")
	}
	if c.APISecret == "" {
		return fmt.Errorf("cloudinary API secret is required")
	}
	return nil
}
