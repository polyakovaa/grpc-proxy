package handler

import (
	"context"

	"github.com/polyakovaa/grpcproxy/event_service/internal/service"
	"github.com/polyakovaa/grpcproxy/gen/event"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EventHandler struct {
	event.UnimplementedEventServiceServer
	eventService *service.EventService
}

func NewEventHandler(eventService *service.EventService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

func (h *EventHandler) CreateEvent(ctx context.Context, req *event.CreateEventRequest) (*event.EventResponse, error) {
	createdEvent, err := h.eventService.CreateEvent(req.Title, req.Description, req.Date, req.OrganizerId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create event: %v", err)
	}

	return &event.EventResponse{
		EventId:     createdEvent.ID,
		Title:       createdEvent.Title,
		Description: createdEvent.Description,
		Date:        createdEvent.Date,
		OrganizerId: createdEvent.OrganizerID,
	}, nil
}

func (h *EventHandler) GetEvent(ctx context.Context, req *event.GetEventRequest) (*event.EventResponse, error) {
	eventFound, err := h.eventService.GetEvent(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "event not found: %v", err)
	}

	return &event.EventResponse{
		EventId:     eventFound.ID,
		Title:       eventFound.Title,
		Description: eventFound.Description,
		Date:        eventFound.Date,
		OrganizerId: eventFound.OrganizerID,
	}, nil
}

func (h *EventHandler) JoinEvent(ctx context.Context, req *event.JoinEventRequest) (*event.JoinEventResponse, error) {
	joinID, err := h.eventService.JoinEvent(req.EventId, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to join event: %v", err)
	}

	return &event.JoinEventResponse{
		Success: true,
		Message: "Successfully joined event",
		JoinId:  joinID,
	}, nil
}

func (h *EventHandler) ListEvents(ctx context.Context, req *event.ListEventsRequest) (*event.ListEventsResponse, error) {
	eventsList, totalCount, err := h.eventService.GetEvents(req.Limit, req.Offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list events: %v", err)
	}
	var events []*event.EventResponse

	for _, evt := range eventsList {
		e := &event.EventResponse{
			EventId:     evt.ID,
			Title:       evt.Title,
			Description: evt.Description,
			Date:        evt.Date,
			OrganizerId: evt.OrganizerID,
		}
		events = append(events, e)
	}
	return &event.ListEventsResponse{
		Events:     events,
		TotalCount: totalCount,
	}, nil
}
