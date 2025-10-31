package cloudinary

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// Service provides Cloudinary operations
type Service struct {
	cld          *cloudinary.Cloudinary
	uploadFolder string
}

// UploadResult contains information about an uploaded file
type UploadResult struct {
	PublicID     string    `json:"public_id"`
	SecureURL    string    `json:"secure_url"`
	URL          string    `json:"url"`
	Format       string    `json:"format"`
	ResourceType string    `json:"resource_type"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Bytes        int       `json:"bytes"`
	CreatedAt    time.Time `json:"created_at"`
}

// UploadOptions provides options for uploading files
type UploadOptions struct {
	Folder           string   // Subfolder within the main upload folder
	PublicID         string   // Custom public ID (optional)
	Overwrite        bool     // Whether to overwrite existing files
	UniqueFilename   bool     // Generate unique filename
	Transformation   string   // Transformation string (e.g., "w_500,h_500,c_fill")
	Tags             []string // Tags to add to the uploaded file
	ResourceType     string   // Resource type: image, video, raw, auto (default: auto)
	AllowedFormats   []string // Allowed file formats
}

// NewService creates a new Cloudinary service
func NewService(config *Config) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid cloudinary config: %w", err)
	}

	fmt.Printf("🔐 NewService - Initializing Cloudinary with CloudName=%s, APIKey=%s (secret hidden), UploadFolder=%s\n", 
		config.CloudName, config.APIKey, config.UploadFolder)
	
	cld, err := cloudinary.NewFromParams(config.CloudName, config.APIKey, config.APISecret)
	if err != nil {
		fmt.Printf("❌ NewService - Failed to initialize Cloudinary: %v\n", err)
		return nil, fmt.Errorf("failed to initialize cloudinary: %w", err)
	}

	fmt.Printf("✅ NewService - Cloudinary initialized successfully\n")
	
	return &Service{
		cld:          cld,
		uploadFolder: config.UploadFolder,
	}, nil
}

// UploadFile uploads a file to Cloudinary
func (s *Service) UploadFile(ctx context.Context, file io.Reader, filename string, opts *UploadOptions) (*UploadResult, error) {
	if opts == nil {
		opts = &UploadOptions{}
	}

	// Build folder path
	folder := s.uploadFolder
	if opts.Folder != "" {
		folder = filepath.Join(folder, opts.Folder)
	}
	// Convert backslashes to forward slashes for Cloudinary (works on all platforms)
	folder = strings.ReplaceAll(folder, "\\", "/")

	// Build upload parameters
	uploadParams := uploader.UploadParams{
		Folder:         folder,
		UniqueFilename: api.Bool(opts.UniqueFilename),
		Overwrite:      api.Bool(opts.Overwrite),
	}

	if opts.PublicID != "" {
		uploadParams.PublicID = opts.PublicID
	}

	if opts.Transformation != "" {
		uploadParams.Transformation = opts.Transformation
	}

	if len(opts.Tags) > 0 {
		uploadParams.Tags = api.CldAPIArray(opts.Tags)
	}

	if opts.ResourceType != "" {
		uploadParams.ResourceType = opts.ResourceType
	} else {
		uploadParams.ResourceType = "auto"
	}

	if len(opts.AllowedFormats) > 0 {
		uploadParams.AllowedFormats = opts.AllowedFormats
	}

	// Debug: print upload params
	fmt.Printf("📤 UploadFile - uploading to folder=%s, resourceType=%s, filename=%s\n", 
		uploadParams.Folder, uploadParams.ResourceType, filename)
	fmt.Printf("🔧 UploadFile - uploadParams: %+v\n", uploadParams)

	// Upload the file
	result, err := s.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		fmt.Printf("❌ UploadFile - Cloudinary API error: %v\n", err)
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Debug: print RAW Cloudinary response
	fmt.Printf("🔍 UploadFile - RAW Cloudinary result: %+v\n", result)
	
	// Check if result is empty/invalid
	if result == nil {
		fmt.Printf("❌ UploadFile - Cloudinary returned nil result\n")
		return nil, fmt.Errorf("cloudinary returned nil result")
	}
	
	if result.PublicID == "" {
		fmt.Printf("❌ UploadFile - Cloudinary returned empty PublicID (possible auth failure)\n")
		fmt.Printf("🔍 UploadFile - Error field: %+v\n", result.Error)
		return nil, fmt.Errorf("cloudinary upload failed: empty public ID (check credentials)")
	}

	// Debug: print Cloudinary response
	fmt.Printf("✅ UploadFile - Cloudinary response: PublicID=%s, URL=%s, Size=%dx%d, Bytes=%d\n", 
		result.PublicID, result.SecureURL, result.Width, result.Height, result.Bytes)

	uploadResult := &UploadResult{
		PublicID:     result.PublicID,
		SecureURL:    result.SecureURL,
		URL:          result.URL,
		Format:       result.Format,
		ResourceType: result.ResourceType,
		Width:        result.Width,
		Height:       result.Height,
		Bytes:        result.Bytes,
		CreatedAt:    result.CreatedAt,
	}

	fmt.Printf("📦 UploadFile - returning result: %+v\n", uploadResult)
	return uploadResult, nil
}

// UploadMultipartFile uploads a multipart file to Cloudinary
func (s *Service) UploadMultipartFile(ctx context.Context, fileHeader *multipart.FileHeader, opts *UploadOptions) (*UploadResult, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open multipart file: %w", err)
	}
	defer file.Close()

	return s.UploadFile(ctx, file, fileHeader.Filename, opts)
}

// DeleteFile deletes a file from Cloudinary by public ID
func (s *Service) DeleteFile(ctx context.Context, publicID string) error {
	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// DeleteFiles deletes multiple files from Cloudinary by public IDs
func (s *Service) DeleteFiles(ctx context.Context, publicIDs []string) error {
	for _, publicID := range publicIDs {
		if err := s.DeleteFile(ctx, publicID); err != nil {
			return err
		}
	}
	return nil
}

// GetTransformedURL generates a transformed URL for an existing image
// Example transformations:
//   - "w_500,h_500,c_fill" - resize to 500x500 with crop
//   - "w_300,h_300,c_thumb,g_face" - 300x300 thumbnail focused on face
//   - "e_grayscale" - convert to grayscale
//   - "q_auto,f_auto" - auto quality and format
func (s *Service) GetTransformedURL(publicID string, transformation string) (string, error) {
	// Generate transformed URL using Cloudinary's URL generation
	asset, err := s.cld.Image(publicID)
	if err != nil {
		return "", fmt.Errorf("failed to generate asset: %w", err)
	}
	
	url, err := asset.String()
	if err != nil {
		return "", fmt.Errorf("failed to generate URL: %w", err)
	}

	// If transformation is provided, insert it into the URL
	if transformation != "" {
		// URL format: https://res.cloudinary.com/{cloud_name}/image/upload/{transformation}/{publicID}
		parts := strings.Split(url, "/upload/")
		if len(parts) == 2 {
			url = parts[0] + "/upload/" + transformation + "/" + parts[1]
		}
	}

	return url, nil
}

// GetOptimizedURL generates an optimized URL with automatic format and quality
func (s *Service) GetOptimizedURL(publicID string, width, height int) (string, error) {
	transformation := fmt.Sprintf("w_%d,h_%d,c_limit,q_auto,f_auto", width, height)
	return s.GetTransformedURL(publicID, transformation)
}

// GetThumbnailURL generates a thumbnail URL
func (s *Service) GetThumbnailURL(publicID string, size int) (string, error) {
	transformation := fmt.Sprintf("w_%d,h_%d,c_fill,q_auto,f_auto", size, size)
	return s.GetTransformedURL(publicID, transformation)
}

// UploadPetImage uploads a pet image with optimized settings
func (s *Service) UploadPetImage(ctx context.Context, file io.Reader, filename string, petID string) (*UploadResult, error) {
	opts := &UploadOptions{
		Folder:         "pets",
		PublicID:       fmt.Sprintf("pet_%s_%d", petID, time.Now().Unix()),
		UniqueFilename: false,
		Overwrite:      false,
		Tags:           []string{"pet", petID},
		ResourceType:   "image",
		AllowedFormats: []string{"jpg", "jpeg", "png", "webp"},
		// Apply basic optimization during upload
		Transformation: "q_auto,f_auto",
	}

	return s.UploadFile(ctx, file, filename, opts)
}

// UploadVendorImage uploads a vendor/service image with optimized settings
func (s *Service) UploadVendorImage(ctx context.Context, file io.Reader, filename string, vendorID string) (*UploadResult, error) {
	opts := &UploadOptions{
		Folder:         "vendors",
		PublicID:       fmt.Sprintf("vendor_%s_%d", vendorID, time.Now().Unix()),
		UniqueFilename: false,
		Overwrite:      false,
		Tags:           []string{"vendor", vendorID},
		ResourceType:   "image",
		AllowedFormats: []string{"jpg", "jpeg", "png", "webp"},
		Transformation: "q_auto,f_auto",
	}

	return s.UploadFile(ctx, file, filename, opts)
}

// UploadUserAvatar uploads a user avatar with optimized settings
func (s *Service) UploadUserAvatar(ctx context.Context, file io.Reader, filename string, userID string) (*UploadResult, error) {
	opts := &UploadOptions{
		Folder:         "avatars",
		PublicID:       fmt.Sprintf("user_%s", userID),
		UniqueFilename: false,
		Overwrite:      true, // Allow overwriting existing avatar
		Tags:           []string{"avatar", userID},
		ResourceType:   "image",
		AllowedFormats: []string{"jpg", "jpeg", "png", "webp"},
		// Apply transformation for avatar: 300x300 circle crop
		Transformation: "w_300,h_300,c_fill,g_face,q_auto,f_auto",
	}

	return s.UploadFile(ctx, file, filename, opts)
}

// UploadServiceImage uploads a service image with optimized settings
func (s *Service) UploadServiceImage(ctx context.Context, file io.Reader, filename string, serviceID string) (*UploadResult, error) {
	opts := &UploadOptions{
		Folder:         "services",
		PublicID:       fmt.Sprintf("service_%s_%d", serviceID, time.Now().Unix()),
		UniqueFilename: false,
		Overwrite:      false,
		Tags:           []string{"service", serviceID},
		ResourceType:   "image",
		AllowedFormats: []string{"jpg", "jpeg", "png", "webp"},
		Transformation: "q_auto,f_auto",
	}

	return s.UploadFile(ctx, file, filename, opts)
}

// ExtractPublicIDFromURL extracts the public ID from a Cloudinary URL
func (s *Service) ExtractPublicIDFromURL(url string) string {
	// URL format: https://res.cloudinary.com/{cloud_name}/image/upload/v{version}/{folder}/{publicID}.{format}
	parts := strings.Split(url, "/upload/")
	if len(parts) < 2 {
		return ""
	}

	// Remove version and get the path
	path := parts[1]
	parts = strings.Split(path, "/")
	if len(parts) < 2 {
		return ""
	}

	// Skip version (v123456) if present
	startIdx := 0
	if strings.HasPrefix(parts[0], "v") {
		startIdx = 1
	}

	// Join remaining parts and remove extension
	publicIDWithExt := strings.Join(parts[startIdx:], "/")
	publicID := strings.TrimSuffix(publicIDWithExt, filepath.Ext(publicIDWithExt))

	return publicID
}
