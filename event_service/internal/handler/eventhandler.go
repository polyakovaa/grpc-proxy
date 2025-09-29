package handler

import (
	"github.com/polyakovaa/grpcproxy/event_service/internal/service"
	"github.com/polyakovaa/grpcproxy/gen/event"
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
