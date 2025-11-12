package command

import (
	"time"
	"whisko-petcare/internal/domain/aggregate"
)

// ============================================
// User Commands
// ============================================

// CreateUser represents a command to create a new user
type CreateUser struct {
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone,omitempty"`
	Address  string `json:"address,omitempty"`
	ImageUrl string `json:"image_url,omitempty"`
}

// UpdateUserProfile represents a command to update user profile
type UpdateUserProfile struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

// UpdateUserContact represents a command to update user contact info
type UpdateUserContact struct {
	UserID  string `json:"user_id"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

// DeleteUser represents a command to delete a user
type DeleteUser struct {
	UserID string `json:"user_id"`
}

// RegisterUser represents a command to register a new user with authentication
type RegisterUser struct {
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// ChangeUserPassword represents a command to change user password
type ChangeUserPassword struct {
	UserID      string `json:"user_id"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// RecordUserLogin represents a command to record user login
type RecordUserLogin struct {
	UserID string `json:"user_id"`
}

// ============================================
// Payment Commands
// ============================================

// CreatePaymentCommand represents a command to create a new payment
type CreatePaymentCommand struct {
	UserID      string                  `json:"user_id"`
	Amount      int                     `json:"amount"` // Amount in VND
	Description string                  `json:"description"`
	Items       []aggregate.PaymentItem `json:"items"`
	// Schedule-related fields for auto-creating schedule after payment
	VendorID    string                  `json:"vendor_id"`
	PetID       string                  `json:"pet_id"`
	ServiceIDs  []string                `json:"service_ids"`
	StartTime   string                  `json:"start_time"` // RFC3339 format
	EndTime     string                  `json:"end_time"`   // RFC3339 format
}

// CreatePaymentResponse represents a payment creation response
type CreatePaymentResponse struct {
	PaymentID   string `json:"payment_id"`
	OrderCode   int64  `json:"order_code"`
	CheckoutURL string `json:"checkout_url"`
	QRCode      string `json:"qr_code"`
	Amount      int    `json:"amount"`
	Status      string `json:"status"`
	ExpiredAt   string `json:"expired_at"`
}

// CancelPaymentCommand represents a command to cancel a payment
type CancelPaymentCommand struct {
	PaymentID string `json:"payment_id"`
	Reason    string `json:"reason"`
}

// ConfirmPaymentCommand represents a command to confirm a payment (typically from webhook)
type ConfirmPaymentCommand struct {
	OrderCode int64 `json:"order_code"`
}

// ============================================
// Pet Commands
// ============================================

// CreatePet represents a command to create a new pet
type CreatePet struct {
	UserID   string  `json:"user_id"`
	Name     string  `json:"name"`
	Species  string  `json:"species"`
	Breed    string  `json:"breed"`
	Age      int     `json:"age"`
	Weight   float64 `json:"weight"`
	ImageUrl string  `json:"image_url,omitempty"`
}

// UpdatePet represents a command to update pet information
type UpdatePet struct {
	PetID   string  `json:"pet_id"`
	Name    string  `json:"name"`
	Species string  `json:"species"`
	Breed   string  `json:"breed"`
	Age     int     `json:"age"`
	Weight  float64 `json:"weight"`
}

// DeletePet represents a command to delete a pet
type DeletePet struct {
	PetID string `json:"pet_id"`
}

// Pet Health Commands
// ============================================

// AddPetVaccination represents a command to add a vaccination record
type AddPetVaccination struct {
	PetID        string    `json:"pet_id"`
	VaccineName  string    `json:"vaccine_name"`
	Date         time.Time `json:"date"`
	NextDueDate  time.Time `json:"next_due_date,omitempty"`
	Veterinarian string    `json:"veterinarian,omitempty"`
	Notes        string    `json:"notes,omitempty"`
}

// AddPetMedicalRecord represents a command to add a medical record
type AddPetMedicalRecord struct {
	PetID        string    `json:"pet_id"`
	Date         time.Time `json:"date"`
	Description  string    `json:"description"`
	Treatment    string    `json:"treatment,omitempty"`
	Veterinarian string    `json:"veterinarian,omitempty"`
	Diagnosis    string    `json:"diagnosis,omitempty"`
	Notes        string    `json:"notes,omitempty"`
}

// AddPetAllergy represents a command to add an allergy
type AddPetAllergy struct {
	PetID         string    `json:"pet_id"`
	Allergen      string    `json:"allergen"`
	Severity      string    `json:"severity"`
	Symptoms      string    `json:"symptoms,omitempty"`
	DiagnosedDate time.Time `json:"diagnosed_date,omitempty"`
	Notes         string    `json:"notes,omitempty"`
}

// RemovePetAllergy represents a command to remove an allergy
type RemovePetAllergy struct {
	PetID     string `json:"pet_id"`
	AllergyID string `json:"allergy_id"`
}

// ============================================
// Vendor Commands
// ============================================

// CreateVendor represents a command to create a new vendor
type CreateVendor struct {
	VendorID string `json:"vendor_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	ImageUrl string `json:"image_url,omitempty"`
}

// UpdateVendor represents a command to update vendor information
type UpdateVendor struct {
	VendorID string `json:"vendor_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

// DeleteVendor represents a command to delete a vendor
type DeleteVendor struct {
	VendorID string `json:"vendor_id"`
}

// UpdateVendorBankAccount represents a command to update vendor's bank account
type UpdateVendorBankAccount struct {
	VendorID      string `json:"vendor_id"`
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
	AccountName   string `json:"account_name"`
	BankBranch    string `json:"bank_branch"`
}

// ============================================
// Service Commands (Vendor Services)
// ============================================

// CreateService represents a command to create a new service
type CreateService struct {
	VendorID    string   `json:"vendor_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       int      `json:"price"`        // Price in VND
	Duration    int      `json:"duration_minutes"` // Duration in minutes
	Tags        []string `json:"tags,omitempty"`
	ImageUrl    string   `json:"image_url,omitempty"`
}

// UpdateService represents a command to update service information
type UpdateService struct {
	ServiceID   string   `json:"service_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       int      `json:"price"`
	Duration    int      `json:"duration_minutes"`
	Tags        []string `json:"tags,omitempty"`
}

// DeleteService represents a command to delete a service
type DeleteService struct {
	ServiceID string `json:"service_id"`
}

// ==================== Schedule Commands ====================

// CreateSchedule represents a command to create a new schedule
type CreateSchedule struct {
	UserID        string              `json:"user_id"`
	VendorID      string              `json:"vendor_id"`
	PetID         string              `json:"pet_id"`
	ServiceIDs    []string            `json:"service_ids"`
	StartTime     string              `json:"start_time"` // RFC3339 format
	EndTime       string              `json:"end_time"`   // RFC3339 format
	BookingUser   BookingUserData     `json:"booking_user"`
	BookedVendor  BookedVendorData    `json:"booked_vendor"`
	AssignedPet   AssignedPetData     `json:"assigned_pet"`
	PaymentID     string              `json:"payment_id,omitempty"`      // For auto-creating payout
	TotalPrice    int                 `json:"total_price,omitempty"`     // For auto-creating payout
}

type BookingUserData struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type BookedVendorData struct {
	Name     string              `json:"name"`
	Location string              `json:"location"`
	Phone    string              `json:"phone"`
	Services []BookedServiceData `json:"services"`
}

type BookedServiceData struct {
	ServiceID string `json:"service_id"`
	Name      string `json:"name"`
}

type AssignedPetData struct {
	Name    string  `json:"name"`
	Species string  `json:"species"`
	Breed   string  `json:"breed"`
	Age     int     `json:"age"`
	Weight  float64 `json:"weight"`
}

// ChangeScheduleStatus represents a command to change schedule status
type ChangeScheduleStatus struct {
	ScheduleID string `json:"schedule_id"`
	Status     string `json:"status"` // pending, confirmed, completed, cancelled
}

// CompleteSchedule represents a command to complete a schedule
type CompleteSchedule struct {
	ScheduleID string `json:"schedule_id"`
}

// CancelSchedule represents a command to cancel a schedule
type CancelSchedule struct {
	ScheduleID string `json:"schedule_id"`
	Reason     string `json:"reason"`
}

// ==================== VendorStaff Commands ====================

// CreateVendorStaff represents a command to create a new vendor staff
// This command will find user by email, create a vendor with default/provided values, and link them
type CreateVendorStaff struct {
	Email         string `json:"email"`           // User email to find/link
	VendorName    string `json:"vendor_name"`     // Vendor name (optional, defaults to user name + "'s Vendor")
	VendorEmail   string `json:"vendor_email"`    // Vendor email (optional, defaults to user email)
	VendorPhone   string `json:"vendor_phone"`    // Vendor phone (optional, defaults to user phone)
	VendorAddress string `json:"vendor_address"`  // Vendor address (optional, defaults to user address)
}

// DeleteVendorStaff represents a command to delete a vendor staff
type DeleteVendorStaff struct {
	UserID   string `json:"user_id"`
	VendorID string `json:"vendor_id"`
}

// ============================================
// Update Image Commands
// ============================================

// UpdateUserImage represents a command to update user image URL
type UpdateUserImage struct {
	UserID   string `json:"user_id"`
	ImageUrl string `json:"image_url"`
}

// UpdatePetImage represents a command to update pet image URL
type UpdatePetImage struct {
	PetID    string `json:"pet_id"`
	ImageUrl string `json:"image_url"`
}

// UpdateVendorImage represents a command to update vendor image URL
type UpdateVendorImage struct {
	VendorID string `json:"vendor_id"`
	ImageUrl string `json:"image_url"`
}

// UpdateServiceImage represents a command to update service image URL
type UpdateServiceImage struct {
	ServiceID string `json:"service_id"`
	ImageUrl  string `json:"image_url"`
}

// ============================================
// Payout Commands
// ============================================

// RequestPayout represents a command to create a new payout request
type RequestPayout struct {
	VendorID   string `json:"vendor_id"`
	PaymentID  string `json:"payment_id"`
	ScheduleID string `json:"schedule_id"`
	Amount     int    `json:"amount"`
	Notes      string `json:"notes,omitempty"`
}
