package aggregate

import (
	"fmt"
	"time"
	"whisko-petcare/internal/domain/event"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserRole represents the role of a user
type UserRole string

const (
	RoleAdmin  UserRole = "Admin"
	RoleVendor UserRole = "Vendor"
	RoleUser   UserRole = "User"
)

// IsValid checks if the role is valid
func (r UserRole) IsValid() bool {
	return r == RoleAdmin || r == RoleVendor || r == RoleUser
}

type User struct {
	id             string
	name           string
	email          string
	phone          string
	address        string
	hashedPassword string
	role           UserRole
	imageUrl       string
	lastLoginAt    *time.Time
	version        int
	createdAt      time.Time
	updatedAt      time.Time
	isActive       bool

	uncommittedEvents []event.DomainEvent
}

func NewUser(id, name, email string, imageUrl ...string) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	user := &User{
		id:        uuid.New().String(),
		name:      name,
		email:     email,
		role:      RoleUser, // Default role
		version:   1,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		isActive:  true,
	}

	// Set imageUrl if provided
	if len(imageUrl) > 0 && imageUrl[0] != "" {
		user.imageUrl = imageUrl[0]
	}

	user.raiseEvent(&event.UserCreated{
		UserID:         user.id,  // Use the generated UUID, not the parameter
		Name:           name,
		Email:          email,
		Phone:          user.phone,
		Address:        user.address,
		HashedPassword: user.hashedPassword, // Empty for non-auth users
		Role:           string(user.role),
		ImageUrl:       user.imageUrl,
		IsActive:       user.isActive,
		Timestamp:      user.createdAt,
	})

	return user, nil
}

// NewUserWithPassword creates a new user with authentication
func NewUserWithPassword(id, name, email, password string) (*User, error) {
	return NewUserWithPasswordAndRole(id, name, email, password, RoleUser)
}

// NewUserWithPasswordAndRole creates a new user with authentication and specific role
func NewUserWithPasswordAndRole(id, name, email, password string, role UserRole) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}
	if !role.IsValid() {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	user := &User{
		id:             uuid.New().String(),
		name:           name,
		email:          email,
		hashedPassword: string(hashedPassword),
		role:           role,
		version:        1,
		createdAt:      now,
		updatedAt:      now,
		isActive:       true,
	}

	user.raiseEvent(&event.UserCreated{
		UserID:         user.id,  // Use the generated UUID, not the parameter
		Name:           name,
		Email:          email,
		Phone:          user.phone,
		Address:        user.address,
		HashedPassword: user.hashedPassword,
		Role:           string(user.role),
		ImageUrl:       user.imageUrl,  // Add ImageUrl field
		IsActive:       user.isActive,
		Timestamp:      user.createdAt,
	})

	return user, nil
}

func NewUserFromHistory(events []event.DomainEvent) (*User, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("no events provided")
	}

	user := &User{}

	for _, e := range events {
		if err := user.applyEvent(e); err != nil {
			return nil, fmt.Errorf("failed to apply event %s: %w", e.EventType(), err)
		}
	}

	return user, nil
}

func (u *User) UpdateProfile(name, email string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	u.raiseEvent(&event.UserProfileUpdated{
		UserID:       u.id,
		Name:         name,
		Email:        email,
		EventVersion: u.version + 1,
		Timestamp:    time.Now(),
	})

	return nil
}

func (u *User) UpdateImageUrl(imageUrl string) error {
	u.raiseEvent(&event.UserImageUpdated{
		UserID:       u.id,
		ImageUrl:     imageUrl,
		EventVersion: u.version + 1,
		Timestamp:    time.Now(),
	})

	return nil
}

func (u *User) UpdateContactInfo(phone, address string) error {
	u.raiseEvent(&event.UserContactUpdated{
		UserID:       u.id,
		Phone:        phone,
		Address:      address,
		EventVersion: u.version + 1,
		Timestamp:    time.Now(),
	})

	return nil
}

func (u *User) Delete() error {
	u.raiseEvent(&event.UserDeleted{
		UserID:       u.id,
		EventVersion: u.version + 1,
		Timestamp:    time.Now(),
	})

	return nil
}

// VerifyPassword verifies if the provided password matches the hashed password
func (u *User) VerifyPassword(password string) error {
	if u.hashedPassword == "" {
		return fmt.Errorf("user has no password set")
	}
	return bcrypt.CompareHashAndPassword([]byte(u.hashedPassword), []byte(password))
}

// ChangePassword changes the user's password
func (u *User) ChangePassword(oldPassword, newPassword string) error {
	// Verify old password
	if err := u.VerifyPassword(oldPassword); err != nil {
		return fmt.Errorf("invalid old password")
	}

	// Validate new password
	if len(newPassword) < 6 {
		return fmt.Errorf("new password must be at least 6 characters")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	u.hashedPassword = string(hashedPassword)
	u.updatedAt = time.Now()
	u.version++

	u.raiseEvent(&event.UserPasswordChanged{
		UserID:         u.id,
		HashedPassword: u.hashedPassword,
		EventVersion:   u.version,
		Timestamp:      u.updatedAt,
	})

	return nil
}

// SetPassword sets a new password (for initial password setup or reset)
func (u *User) SetPassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	u.hashedPassword = string(hashedPassword)
	u.updatedAt = time.Now()
	u.version++

	u.raiseEvent(&event.UserPasswordChanged{
		UserID:         u.id,
		HashedPassword: u.hashedPassword,
		EventVersion:   u.version,
		Timestamp:      u.updatedAt,
	})

	return nil
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.lastLoginAt = &now
	u.updatedAt = now
	u.version++

	u.raiseEvent(&event.UserLoggedIn{
		UserID:       u.id,
		EventVersion: u.version,
		Timestamp:    now,
	})
}

// SetHashedPassword sets the hashed password directly (for repository reconstruction)
func (u *User) SetHashedPassword(hashedPassword string) {
	u.hashedPassword = hashedPassword
}

// Deactivate deactivates the user account
func (u *User) Deactivate() {
	u.isActive = false
	u.updatedAt = time.Now()
}

// Activate activates the user account
func (u *User) Activate() {
	u.isActive = true
	u.updatedAt = time.Now()
}

// UpdateRole updates the user's role (admin only)
func (u *User) UpdateRole(newRole UserRole) error {
	if !newRole.IsValid() {
		return fmt.Errorf("invalid role: %s", newRole)
	}
	u.role = newRole
	u.updatedAt = time.Now()
	u.version++

	u.raiseEvent(&event.UserRoleUpdated{
		UserID:       u.id,
		Role:         string(u.role),
		EventVersion: u.version,
		Timestamp:    u.updatedAt,
	})

	return nil
}

// HasRole checks if user has a specific role
func (u *User) HasRole(role UserRole) bool {
	return u.role == role
}

// IsAdmin checks if user is an admin
func (u *User) IsAdmin() bool {
	return u.role == RoleAdmin
}

// IsVendor checks if user is a vendor
func (u *User) IsVendor() bool {
	return u.role == RoleVendor
}

func (u *User) GetUncommittedEvents() []event.DomainEvent {
	return u.uncommittedEvents
}

func (u *User) ClearUncommittedEvents() {
	u.uncommittedEvents = nil
}

func (u *User) raiseEvent(ev event.DomainEvent) {
	u.uncommittedEvents = append(u.uncommittedEvents, ev)
	u.applyEvent(ev)
}

func (u *User) applyEvent(ev event.DomainEvent) error {
	switch e := ev.(type) {
	case *event.UserCreated:
		u.id = e.UserID
		u.name = e.Name
		u.email = e.Email
		u.phone = e.Phone
		u.address = e.Address
		u.version = 1
		u.createdAt = e.Timestamp
		u.updatedAt = e.Timestamp

	case *event.UserProfileUpdated:
		u.name = e.Name
		u.email = e.Email
		u.version = e.EventVersion
		u.updatedAt = e.Timestamp

	case *event.UserContactUpdated:
		u.phone = e.Phone
		u.address = e.Address
		u.version = e.EventVersion
		u.updatedAt = e.Timestamp

	case *event.UserImageUpdated:
		u.imageUrl = e.ImageUrl
		u.version = e.EventVersion
		u.updatedAt = e.Timestamp

	case *event.UserDeleted:
		u.version = e.EventVersion
		u.updatedAt = e.Timestamp
		u.isActive = false

	default:
		return fmt.Errorf("unknown event type: %T", ev)
	}

	return nil
}

// Getters
func (u *User) ID() string              { return u.id }
func (u *User) Name() string            { return u.name }
func (u *User) Email() string           { return u.email }
func (u *User) Phone() string           { return u.phone }
func (u *User) Address() string         { return u.address }
func (u *User) HashedPassword() string  { return u.hashedPassword }
func (u *User) Role() UserRole          { return u.role }
func (u *User) ImageUrl() string        { return u.imageUrl }
func (u *User) LastLoginAt() *time.Time { return u.lastLoginAt }
func (u *User) Version() int            { return u.version }
func (u *User) CreatedAt() time.Time    { return u.createdAt }
func (u *User) UpdatedAt() time.Time    { return u.updatedAt }
func (u *User) IsActive() bool          { return u.isActive }

// Entity interface implementation
func (u *User) GetID() string    { return u.id }
func (u *User) SetID(id string)  { u.id = id }
func (u *User) GetVersion() int  { return u.version }
func (u *User) SetVersion(v int) { u.version = v }	

// AggregateRoot interface implementation
func (u *User) MarkEventsAsCommitted() {
	u.uncommittedEvents = nil
}

func (u *User) LoadFromHistory(events []event.DomainEvent) error {
	for _, e := range events {
		if err := u.applyEvent(e); err != nil {
			return fmt.Errorf("failed to apply event %s: %w", e.EventType(), err)
		}
	}
	return nil
	}
