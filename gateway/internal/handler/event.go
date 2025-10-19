package handler

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/polyakovaa/grpcproxy/gateway/internal/utils"
	"github.com/polyakovaa/grpcproxy/gen/auth"
	"github.com/polyakovaa/grpcproxy/gen/event"
)

type EventHandler struct {
	eventClient event.EventServiceClient
	authClient  auth.AuthServiceClient
}

func NewEventHandler(eventClient event.EventServiceClient, authClient auth.AuthServiceClient) *EventHandler {
	return &EventHandler{
		eventClient: eventClient,
		authClient:  authClient,
	}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	if h.eventClient == nil {
		c.JSON(503, gin.H{"error": "Event service unavailable"})
		return
	}

	var request struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Date        string `json:"date"`
	}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	userID, err := h.getAuthenticatedUserID(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	response, err := h.eventClient.CreateEvent(c.Request.Context(), &event.CreateEventRequest{
		Title:       request.Title,
		Description: request.Description,
		Date:        request.Date,
		OrganizerId: userID,
	})
	if err != nil {
		log.Printf("CreateEvent error: %v", err)
		utils.HandleGRPCError(c, err)
		return
	}

	c.JSON(201, gin.H{
		"id":          response.EventId,
		"title":       response.Title,
		"description": response.Description,
		"date":        response.Date,
	})
}

func (h *EventHandler) GetEvent(c *gin.Context) {
	if h.eventClient == nil {
		c.JSON(503, gin.H{"error": "Event service unavailable"})
		return
	}

	eventID := c.Param("id")

	response, err := h.eventClient.GetEvent(c.Request.Context(), &event.GetEventRequest{
		EventId: eventID,
	})

	if err != nil {
		log.Printf("GetEvent error: %v", err)
		utils.HandleGRPCError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"id":          response.EventId,
		"title":       response.Title,
		"description": response.Description,
		"date":        response.Date,
	})

}

func (h *EventHandler) GetEvents(c *gin.Context) {
	if h.eventClient == nil {
		c.JSON(503, gin.H{"error": "Event service unavailable"})
		return
	}

	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "10")

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid offset"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		c.JSON(400, gin.H{"error": "invalid limit"})
		return
	}

	resp, err := h.eventClient.ListEvents(c.Request.Context(), &event.ListEventsRequest{
		Offset: int32(offset),
		Limit:  int32(limit),
	})
	if err != nil {
		log.Printf("ListEvents error: %v", err)
		utils.HandleGRPCError(c, err)
		return
	}

	events := []gin.H{}
	for _, e := range resp.Events {
		events = append(events, gin.H{
			"id":          e.EventId,
			"title":       e.Title,
			"description": e.Description,
			"date":        e.Date,
			"organizerId": e.OrganizerId,
		})
	}

	c.JSON(200, gin.H{
		"events":      events,
		"total_count": resp.TotalCount,
	})

}

func (h *EventHandler) JoinEvent(c *gin.Context) {
	if h.eventClient == nil {
		c.JSON(503, gin.H{"error": "Event service unavailable"})
		return
	}

	eventID := c.Param("id")

	userID, err := h.getAuthenticatedUserID(c)

	if err != nil {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	response, err := h.eventClient.JoinEvent(c.Request.Context(), &event.JoinEventRequest{
		EventId: eventID,
		UserId:  userID,
	})

	if err != nil {
		utils.HandleGRPCError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"success": response.Success,
		"message": response.Message,
		"join_id": response.JoinId,
	})
}

func (h *EventHandler) getAuthenticatedUserID(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return "", fmt.Errorf("invalid authorization format")
	}

	resp, err := h.authClient.ValidateToken(c.Request.Context(), &auth.ValidateTokenRequest{
		Token: tokenString,
	})
	if err != nil {
		return "", fmt.Errorf("invalid token: %v", err)
	}
	return resp.UserId, nil
}
