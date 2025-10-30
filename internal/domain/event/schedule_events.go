package event

import "time"

// Embedded structs for schedule event data
type BookingUserData struct {
	UserID  string `json:"user_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type BookedVendorData struct {
	ShopID         string               `json:"shop_id"`
	Name           string               `json:"name"`
	Location       string               `json:"location"`
	Phone          string               `json:"phone"`
	BookedServices []BookedServicesData `json:"booked_services"`
}

type BookedServicesData struct {
	ServiceID string `json:"service_id"`
	Name      string `json:"name"`
}

type PetAssignedData struct {
	PetID   string  `json:"pet_id"`
	Name    string  `json:"name"`
	Species string  `json:"species"`
	Breed   string  `json:"breed"`
	Age     int     `json:"age"`
	Weight  float64 `json:"weight"`
}

// ScheduleCreated event
type ScheduleCreated struct {
	ScheduleID   string           `json:"schedule_id"`
	BookingUser  BookingUserData  `json:"booking_user"`
	BookedVendor BookedVendorData `json:"booked_vendor"`
	AssignedPet  PetAssignedData  `json:"assigned_pet"`
	StartTime    time.Time        `json:"start_time"`
	EndTime      time.Time        `json:"end_time"`
	Status       string           `json:"status"`
	Timestamp    time.Time        `json:"timestamp"`
}

func (e *ScheduleCreated) EventType() string     { return "ScheduleCreated" }
func (e *ScheduleCreated) AggregateID() string   { return e.ScheduleID }
func (e *ScheduleCreated) OccurredAt() time.Time { return e.Timestamp }
func (e *ScheduleCreated) Version() int          { return 1 }

// ScheduleStatusChanged event
type ScheduleStatusChanged struct {
	ScheduleID   string    `json:"schedule_id"`
	OldStatus    string    `json:"old_status"`
	NewStatus    string    `json:"new_status"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *ScheduleStatusChanged) EventType() string     { return "ScheduleStatusChanged" }
func (e *ScheduleStatusChanged) AggregateID() string   { return e.ScheduleID }
func (e *ScheduleStatusChanged) OccurredAt() time.Time { return e.Timestamp }
func (e *ScheduleStatusChanged) Version() int          { return e.EventVersion }

// ScheduleCancelled event
type ScheduleCancelled struct {
	ScheduleID   string    `json:"schedule_id"`
	Reason       string    `json:"reason"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *ScheduleCancelled) EventType() string     { return "ScheduleCancelled" }
func (e *ScheduleCancelled) AggregateID() string   { return e.ScheduleID }
func (e *ScheduleCancelled) OccurredAt() time.Time { return e.Timestamp }
func (e *ScheduleCancelled) Version() int          { return e.EventVersion }

// ScheduleCompleted event
type ScheduleCompleted struct {
	ScheduleID   string    `json:"schedule_id"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *ScheduleCompleted) EventType() string     { return "ScheduleCompleted" }
func (e *ScheduleCompleted) AggregateID() string   { return e.ScheduleID }
func (e *ScheduleCompleted) OccurredAt() time.Time { return e.Timestamp }
func (e *ScheduleCompleted) Version() int          { return e.EventVersion }
