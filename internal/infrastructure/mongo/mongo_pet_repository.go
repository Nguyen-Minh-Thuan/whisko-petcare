package mongo

import (
	"context"
	"fmt"
	"time"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/event"

	// "whisko-petcare/internal/domain/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	// "whisko-petcare/internal/infrastructure/mongo/collections"
	// "whisko-petcare/pkg/errors"
)

type MongoPetRepository struct {
	database         *mongo.Database
	entityCollection *mongo.Collection
	eventCollection  *mongo.Collection
	session          mongo.Session
}

func NewMongoPetRepository(database *mongo.Database) *MongoPetRepository {
	return &MongoPetRepository{
		database:         database,
		entityCollection: database.Collection("pets"),
		eventCollection:  database.Collection("pet_events"),
	}
}	

func (r *MongoPetRepository) SetTransaction(tx interface{}) {
	if session, ok := tx.(mongo.Session); ok {
		r.session = session
	}
}

func (r *MongoPetRepository) GetTransaction() interface{} {
	return r.session
}	

func (r *MongoPetRepository) IsTransactional() bool {
	return r.session != nil
}

func (r *MongoPetRepository) Save(ctx context.Context, pet *aggregate.Pet) error {
	// Use session context if in transaction
	var ctxToUse context.Context = ctx
	if r.session != nil {
		ctxToUse = mongo.NewSessionContext(ctx, r.session)
	}

	// Prepare entity document for MongoDB
	entityDoc := bson.M{
		"_id":                  pet.GetID(),
		"version":              pet.GetVersion(),
		"user_id":              pet.UserID(),
		"name":                 pet.Name(),
		"species":              pet.Species(),
		"breed":                pet.Breed(),
		"age":                  pet.Age(),
		"weight":               pet.Weight(),
		"image_url":            pet.ImageUrl(),
		"is_active":            pet.IsActive(),
		"created_at":           pet.CreatedAt(),
		"updated_at":           pet.UpdatedAt(),
		"vaccination_records":  pet.VaccinationRecords(),
		"medical_history":      pet.MedicalHistory(),
		"allergies":            pet.Allergies(),
	}
	
	// Upsert the entity document in the database
	opts := options.Update().SetUpsert(true)
	_, err := r.entityCollection.UpdateOne(ctxToUse, bson.M{"_id": pet.GetID()}, bson.M{"$set": entityDoc}, opts)
	if err != nil {
		return fmt.Errorf("failed to save pet entity: %w", err)
	}

	// Save domain events
	events := pet.GetUncommittedEvents()
	if len(events) > 0 {
		var eventDocs []interface{}
		for _, e := range events {
			eventDoc := bson.M{
				"aggregate_id":  e.AggregateID(),
				"event_type":    e.EventType(),
				"event_version": e.Version(),
				"occurred_at":   e.OccurredAt(),
				"event_data":    e,
			}
			eventDocs = append(eventDocs, eventDoc)
		}

		_, err = r.eventCollection.InsertMany(ctxToUse, eventDocs)
		if err != nil {
			return fmt.Errorf("failed to save pet events: %w", err)
		}

		// Mark events as committed
		pet.MarkEventsAsCommitted()
	}

	return nil
}


func (r *MongoPetRepository) GetByID(ctx context.Context, id string) (*aggregate.Pet, error) {
	var ctxToUse context.Context = ctx
	if r.session != nil {
		ctxToUse = mongo.NewSessionContext(ctx, r.session)
	}
	
	// Retrieve the pet entity document
	var petDoc bson.M
	err := r.entityCollection.FindOne(ctxToUse, bson.M{"_id": id}).Decode(&petDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("pet not found: %w", err)
		}
		return nil, fmt.Errorf("failed to retrieve pet: %w", err)
	}

	// Reconstruct the Pet aggregate using all required parameters
	imageUrl := getPetString(petDoc, "image_url")
	pet, err := aggregate.NewPet(
		getPetString(petDoc, "user_id"),
		getPetString(petDoc, "name"),
		getPetString(petDoc, "species"),
		getPetString(petDoc, "breed"),
		getPetInt(petDoc, "age"),
		getPetFloat64(petDoc, "weight"),
		imageUrl,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pet aggregate: %w", err)
	}

	// Set version from database
	pet.SetVersion(getPetInt(petDoc, "version"))

	// Reconstruct health data from database
	pet.SetVaccinationRecords(getPetVaccinationRecords(petDoc))
	pet.SetMedicalHistory(getPetMedicalHistory(petDoc))
	pet.SetAllergies(getPetAllergies(petDoc))

	return pet, nil
}

func (r *MongoPetRepository) SaveEvents(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error {
	var ctxToUse context.Context = ctx
	if r.session != nil {
		ctxToUse = mongo.NewSessionContext(ctx, r.session)
	}
	// Create event documents
	var eventDocs []interface{}
	currentVersion := expectedVersion
	for _, evt := range events {
		currentVersion++
		eventDoc := bson.M{
			"aggregate_id": aggregateID,
			"event_type":   evt.EventType(),
			"event_data":   evt,
			"version":      currentVersion,
		}
		eventDocs = append(eventDocs, eventDoc)
	}
	_, err := r.eventCollection.InsertMany(ctxToUse, eventDocs)
	if err != nil {
		return fmt.Errorf("failed to save events to MongoDB: %w", err)
	}
	return nil
}

func (r *MongoPetRepository) GetEvents(ctx context.Context, aggregateID string) ([]event.DomainEvent, error) {
	// Simple implementation that returns empty slice
	// Full event sourcing reconstruction would be more complex
	return []event.DomainEvent{}, nil
}

func (r *MongoPetRepository) GetEventsSince(ctx context.Context, aggregateID string, version int) ([]event.DomainEvent, error) {
	return r.GetEvents(ctx, aggregateID)
}

func (r *MongoPetRepository) GetAllEvents(ctx context.Context) ([]event.DomainEvent, error) {
	return []event.DomainEvent{}, nil
}

// Helper functions for Pet repository
func getPetString(doc bson.M, key string) string {
	if val, ok := doc[key].(string); ok {
		return val
	}
	return ""
}

func getPetInt(doc bson.M, key string) int {
	if val, ok := doc[key].(int32); ok {
		return int(val)
	}
	if val, ok := doc[key].(int); ok {
		return val
	}
	if val, ok := doc[key].(int64); ok {
		return int(val)
	}
	if val, ok := doc[key].(float64); ok {
		return int(val)
	}
	return 0
}

func getPetFloat64(doc bson.M, key string) float64 {
	if val, ok := doc[key].(float64); ok {
		return val
	}
	if val, ok := doc[key].(float32); ok {
		return float64(val)
	}
	if val, ok := doc[key].(int); ok {
		return float64(val)
	}
	if val, ok := doc[key].(int32); ok {
		return float64(val)
	}
	if val, ok := doc[key].(int64); ok {
		return float64(val)
	}
	return 0.0
}

func getPetVaccinationRecords(doc bson.M) []event.VaccinationRecord {
	records := []event.VaccinationRecord{}
	if val, ok := doc["vaccination_records"].(bson.A); ok {
		for _, item := range val {
			if recordMap, ok := item.(bson.M); ok {
				record := event.VaccinationRecord{
					ID:            getPetString(recordMap, "id"),
					VaccineName:   getPetString(recordMap, "vaccine_name"),
					Date:          getPetTime(recordMap, "date"),
					NextDueDate:   getPetTime(recordMap, "next_due_date"),
					Veterinarian:  getPetString(recordMap, "veterinarian"),
					Notes:         getPetString(recordMap, "notes"),
				}
				records = append(records, record)
			}
		}
	}
	return records
}

func getPetMedicalHistory(doc bson.M) []event.MedicalRecord {
	records := []event.MedicalRecord{}
	if val, ok := doc["medical_history"].(bson.A); ok {
		for _, item := range val {
			if recordMap, ok := item.(bson.M); ok {
				record := event.MedicalRecord{
					ID:            getPetString(recordMap, "id"),
					Date:          getPetTime(recordMap, "date"),
					Description:   getPetString(recordMap, "description"),
					Treatment:     getPetString(recordMap, "treatment"),
					Veterinarian:  getPetString(recordMap, "veterinarian"),
					Diagnosis:     getPetString(recordMap, "diagnosis"),
					Notes:         getPetString(recordMap, "notes"),
				}
				records = append(records, record)
			}
		}
	}
	return records
}

func getPetAllergies(doc bson.M) []event.Allergy {
	allergies := []event.Allergy{}
	if val, ok := doc["allergies"].(bson.A); ok {
		for _, item := range val {
			if allergyMap, ok := item.(bson.M); ok {
				allergy := event.Allergy{
					ID:            getPetString(allergyMap, "id"),
					Allergen:      getPetString(allergyMap, "allergen"),
					Severity:      getPetString(allergyMap, "severity"),
					Symptoms:      getPetString(allergyMap, "symptoms"),
					DiagnosedDate: getPetTime(allergyMap, "diagnosed_date"),
					Notes:         getPetString(allergyMap, "notes"),
				}
				allergies = append(allergies, allergy)
			}
		}
	}
	return allergies
}

func getPetTime(doc bson.M, key string) time.Time {
	if val, ok := doc[key].(time.Time); ok {
		return val
	}
	return time.Time{}
}

