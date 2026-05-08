package rewriter

import "encoding/json"

type MessagesRequest struct {
	Metadata *struct {
		UserID string `json:"user_id"`
	} `json:"metadata"`
	Messages []json.RawMessage `json:"messages"`
	System   json.RawMessage  `json:"system"`
}

type Message struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type EventPayload struct {
	Events []Event `json:"events"`
}

type Event struct {
	EventType string    `json:"event_type"`
	EventData EventData `json:"event_data"`
}

type EventData struct {
	DeviceID          string `json:"device_id"`
	Email             string `json:"email"`
	EventName         string `json:"event_name"`
	Env               any    `json:"env,omitempty"`
	Process           any    `json:"process,omitempty"`
	BaseURL           string `json:"baseUrl,omitempty"`
	BaseURLUnderscore string `json:"base_url,omitempty"`
	Gateway           string `json:"gateway,omitempty"`
	AdditionalMeta    string `json:"additional_metadata,omitempty"`
}
