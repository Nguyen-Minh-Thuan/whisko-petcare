package event

import "time"

// ScheduleCreated event
type ScheduleCreated struct {
	ScheduleID string    `json:"schedule_id"`
	UserID     string    `json:"user_id"`
	ShopID     string    `json:"shop_id"`
	PetID      string    `json:"pet_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Status     string    `json:"status"`
	Timestamp  time.Time `json:"timestamp"`
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
