package aggregate

import (
	"fmt"
	"time"
	"whisko-petcare/internal/domain/event"

	"github.com/google/uuid"
)

type PetAssigned struct {
	PetID   string  `json:"pet_id"`
	Name    string  `json:"name"`
	Species string  `json:"species"`
	Breed   string  `json:"breed"`
	Age     int     `json:"age"`
	Weight  float64 `json:"weight"`
}

type BookingUser struct {
	UserID  string `json:"user_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type BookedVendor struct {
	ShopID         string           `json:"shop_id"`
	Name           string           `json:"name"`
	Location       string           `json:"location"`
	Phone          string           `json:"phone"`
	BookedServices []BookedServices `json:"booked_services"`
}

type BookedServices struct {
	ServiceID string `json:"service_id"`
	Name      string `json:"name"`
}

type ScheduleStatus string

const (
	ScheduleStatusPending   ScheduleStatus = "pending"
	ScheduleStatusConfirmed  ScheduleStatus = "confirmed"
	ScheduleStatusCompleted  ScheduleStatus = "completed"
	ScheduleStatusCancelled  ScheduleStatus = "cancelled"
)

type Schedule struct {
	id               string
	bookingUser      BookingUser
	bookedShop       BookedVendor
	assignedPet      PetAssigned
	startTime        time.Time
	endTime          time.Time
	status           ScheduleStatus
	createdAt        time.Time
	updatedAt        time.Time
	version          int
	isActive         bool

	uncommittedEvents []event.DomainEvent
}

func NewSchedule(bookingUser BookingUser, bookedShop BookedVendor, assignedPet PetAssigned, startTime, endTime time.Time) (*Schedule, error) {
	if bookingUser.UserID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}
	if bookedShop.ShopID == "" {
		return nil, fmt.Errorf("shopID cannot be empty")
	}
	if assignedPet.PetID == "" {
		return nil, fmt.Errorf("petID cannot be empty")
	}
	if startTime.IsZero() {
		return nil, fmt.Errorf("startTime cannot be empty")
	}
	if endTime.IsZero() {
		return nil, fmt.Errorf("endTime cannot be empty")
	}
	if endTime.Before(startTime) {
		return nil, fmt.Errorf("endTime must be after startTime")
	}

	schedule := &Schedule{
		id:          uuid.New().String(),
		bookingUser: bookingUser,
		bookedShop:  bookedShop,
		assignedPet: assignedPet,
		startTime:   startTime,
		endTime:     endTime,
		status:      ScheduleStatusPending,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
		version:     1,
		isActive:    true,
	}

	schedule.raiseEvent(&event.ScheduleCreated{
		ScheduleID: schedule.id,
		UserID:     bookingUser.UserID,
		ShopID:     bookedShop.ShopID,
		PetID:      assignedPet.PetID,
		StartTime:  startTime,
		EndTime:    endTime,
		Status:     string(schedule.status),
		Timestamp:  schedule.createdAt,
	})

	return schedule, nil
}

func NewScheduleFromHistory(events []event.DomainEvent) (*Schedule, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("no events provided")
	}

	schedule := &Schedule{}
	for _, e := range events {
		if err := schedule.applyEvent(e); err != nil {
			return nil, fmt.Errorf("failed to apply event %s: %w", e.EventType(), err)
		}
	}

	return schedule, nil
}

// ChangeStatus changes the status of the schedule
func (s *Schedule) ChangeStatus(newStatus ScheduleStatus) error {
	if s.status == newStatus {
		return fmt.Errorf("schedule is already in status %s", newStatus)
	}
	
	s.raiseEvent(&event.ScheduleStatusChanged{
		ScheduleID:   s.id,
		OldStatus:    string(s.status),
		NewStatus:    string(newStatus),
		EventVersion: s.version + 1,
		Timestamp:    time.Now(),
	})

	return nil
}

// Complete marks the schedule as completed
func (s *Schedule) Complete() error {
	if s.status == ScheduleStatusCompleted {
		return fmt.Errorf("schedule is already completed")
	}
	
	s.raiseEvent(&event.ScheduleCompleted{
		ScheduleID:   s.id,
		EventVersion: s.version + 1,
		Timestamp:    time.Now(),
	})
	
	return nil
}

// Cancel cancels the schedule
func (s *Schedule) Cancel(reason string) error {
	if s.status == ScheduleStatusCancelled {
		return fmt.Errorf("schedule is already cancelled")
	}
	
	s.raiseEvent(&event.ScheduleCancelled{
		ScheduleID:   s.id,
		Reason:       reason,
		EventVersion: s.version + 1,
		Timestamp:    time.Now(),
	})
	
	return nil
}

func (s *Schedule) GetUncommittedEvents() []event.DomainEvent {
	return s.uncommittedEvents
}

func (s *Schedule) ClearUncommittedEvents() {
	s.uncommittedEvents = nil
}

func (s *Schedule) raiseEvent(ev event.DomainEvent) {
	s.uncommittedEvents = append(s.uncommittedEvents, ev)
	s.applyEvent(ev)
}
func (s *Schedule) applyEvent(ev event.DomainEvent) error {
	switch e := ev.(type) {
	case *event.ScheduleCreated:
		s.id = e.ScheduleID
		s.bookingUser = BookingUser{
			UserID: e.UserID,
		}
		s.bookedShop = BookedVendor{
			ShopID: e.ShopID,
		}
		s.assignedPet = PetAssigned{
			PetID: e.PetID,
		}
		s.startTime = e.StartTime
		s.endTime = e.EndTime
		s.status = ScheduleStatus(e.Status)
		s.createdAt = e.Timestamp
		s.updatedAt = e.Timestamp
		s.version = 1
		s.isActive = true
		
	case *event.ScheduleStatusChanged:
		s.status = ScheduleStatus(e.NewStatus)
		s.version = e.EventVersion
		s.updatedAt = e.Timestamp
		
	case *event.ScheduleCompleted:
		s.status = ScheduleStatusCompleted
		s.version = e.EventVersion
		s.updatedAt = e.Timestamp
		
	case *event.ScheduleCancelled:
		s.status = ScheduleStatusCancelled
		s.version = e.EventVersion
		s.updatedAt = e.Timestamp
		s.isActive = false
		
	default:
		return fmt.Errorf("unknown event type: %T", ev)
	}
	
	return nil
}

// Getters
func (s *Schedule) ID() string              { return s.id }
func (s *Schedule) BookingUser() BookingUser { return s.bookingUser }
func (s *Schedule) BookedShop() BookedVendor  { return s.bookedShop }
func (s *Schedule) AssignedPet() PetAssigned { return s.assignedPet }
func (s *Schedule) StartTime() time.Time    { return s.startTime }
func (s *Schedule) EndTime() time.Time      { return s.endTime }
func (s *Schedule) Status() ScheduleStatus  { return s.status }
func (s *Schedule) CreatedAt() time.Time    { return s.createdAt }
func (s *Schedule) UpdatedAt() time.Time    { return s.updatedAt }
func (s *Schedule) Version() int            { return s.version }
func (s *Schedule) IsActive() bool          { return s.isActive }

// Entity interface implementation
func (s *Schedule) GetID() string    { return s.id }
func (s *Schedule) GetVersion() int  { return s.version }
func (s *Schedule) SetVersion(v int) { s.version = v }


// AggregateRoot interface implementation
func (s *Schedule) MarkEventsAsCommitted() {
	s.uncommittedEvents = nil
}

func (s *Schedule) LoadFromHistory(events []event.DomainEvent) error {
	for _, e := range events {
		if err := s.applyEvent(e); err != nil {
			return fmt.Errorf("failed to apply event %s: %w", e.EventType(), err)
		}
	}
	return nil
}

