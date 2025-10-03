package repository

import (
	"database/sql"
	"fmt"

	"github.com/polyakovaa/grpcproxy/event_service/internal/model"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{
		db: db,
	}
}

func (r *EventRepository) CreateEvent(event *model.Event) error {
	query := `
		INSERT INTO events (id, title, description, date, organizer_id)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(query,
		event.ID,
		event.Title,
		event.Description,
		event.Date,
		event.OrganizerID,
	)

	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

func (r *EventRepository) AddParticipant(eventID, userID, joinID string) error {
	query := `
		INSERT INTO event_participants (id, event_id, user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (event_id, user_id) DO NOTHING
	`

	_, err := r.db.Exec(query, joinID, eventID, userID)
	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}

func (r *EventRepository) GetEventByID(eventID string) (*model.Event, error) {
	query := `
		SELECT id, title, description, date, organizer_id
		FROM events 
		WHERE id = $1
	`

	var event model.Event
	err := r.db.QueryRow(query, eventID).Scan(
		&event.ID,
		&event.Title,
		&event.Description,
		&event.Date,
		&event.OrganizerID,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("event not found: %s", eventID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return &event, nil
}

func (r *EventRepository) GetEvents(limit, offset int) ([]*model.Event, error) {
	query := `
		SELECT id, title, description, date, organizer_id
		FROM events  
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer rows.Close()

	var events []*model.Event
	for rows.Next() {
		var event model.Event
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.Date,
			&event.OrganizerID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, &event)
	}

	return events, nil
}
