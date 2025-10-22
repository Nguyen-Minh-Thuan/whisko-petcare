package command

import "whisko-petcare/internal/domain/aggregate"

// ============================================
// User Commands
// ============================================

// CreateUser represents a command to create a new user
type CreateUser struct {
	UserID  string `json:"user_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`
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

// ============================================
// Payment Commands
// ============================================

// CreatePaymentCommand represents a command to create a new payment
type CreatePaymentCommand struct {
	UserID      string                  `json:"user_id"`
	Amount      int                     `json:"amount"` // Amount in VND
	Description string                  `json:"description"`
	Items       []aggregate.PaymentItem `json:"items"`
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
	UserID  string  `json:"user_id"`
	Name    string  `json:"name"`
	Species string  `json:"species"`
	Breed   string  `json:"breed"`
	Age     int     `json:"age"`
	Weight  float64 `json:"weight"`
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

// ============================================
// Vendor Commands
// ============================================

// CreateVendor represents a command to create a new vendor
type CreateVendor struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
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

// ============================================
// Service Commands (Vendor Services)
// ============================================

// CreateService represents a command to create a new service
type CreateService struct {
	VendorID    string `json:"vendor_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`        // Price in VND
	Duration    int    `json:"duration_minutes"` // Duration in minutes
}

// UpdateService represents a command to update service information
type UpdateService struct {
	ServiceID   string `json:"service_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	Duration    int    `json:"duration_minutes"`
}

// DeleteService represents a command to delete a service
type DeleteService struct {
	ServiceID string `json:"service_id"`
}

// ==================== Schedule Commands ====================

// CreateSchedule represents a command to create a new schedule
type CreateSchedule struct {
	UserID        string              `json:"user_id"`
	ShopID        string              `json:"shop_id"`
	PetID         string              `json:"pet_id"`
	ServiceIDs    []string            `json:"service_ids"`
	StartTime     string              `json:"start_time"` // RFC3339 format
	EndTime       string              `json:"end_time"`   // RFC3339 format
	BookingUser   BookingUserData     `json:"booking_user"`
	BookedShop    BookedShopData      `json:"booked_shop"`
	AssignedPet   AssignedPetData     `json:"assigned_pet"`
}

type BookingUserData struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type BookedShopData struct {
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
type CreateVendorStaff struct {
	UserID   string `json:"user_id"`
	VendorID string `json:"vendor_id"`
}

// DeleteVendorStaff represents a command to delete a vendor staff
type DeleteVendorStaff struct {
	UserID   string `json:"user_id"`
	VendorID string `json:"vendor_id"`
}
