package model

type Event struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Date        string `json:"date"`
	OrganizerID string `json:"organizer_id"`
}

type EventParticipant struct {
	ID      string `json:"id"`
	EventID string `json:"event_id"`
	UserID  string `json:"user_id"`
}
