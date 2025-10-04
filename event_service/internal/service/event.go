package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/polyakovaa/grpcproxy/event_service/internal/model"
	"github.com/polyakovaa/grpcproxy/event_service/internal/repository"
)

type EventService struct {
	eventRepo *repository.EventRepository
}

func NewEventService(eventRepo *repository.EventRepository) *EventService {
	return &EventService{
		eventRepo: eventRepo,
	}
}

func (s *EventService) CreateEvent(title, description, date, organizerID string) (*model.Event, error) {
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	event := &model.Event{
		ID:          generateID(),
		Title:       title,
		Description: description,
		Date:        date,
		OrganizerID: organizerID,
	}

	err := s.eventRepo.CreateEvent(event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) GetEvent(eventID string) (*model.Event, error) {

	event, err := s.eventRepo.GetEventByID(eventID)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) JoinEvent(eventID, userID string) (string, error) {
	_, err := s.GetEvent(eventID)
	if err != nil {
		return "", fmt.Errorf("event not found")
	}

	joinID := generateID()
	err = s.eventRepo.AddParticipant(eventID, userID, joinID)
	if err != nil {
		return "", err
	}

	return joinID, nil
}

func (s *EventService) GetEvents(limit, offset int32) ([]*model.Event, int32, error) {
	events, count, err := s.eventRepo.GetEvents(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return events, count, nil
}

func generateID() string {
	return uuid.NewString()
}
