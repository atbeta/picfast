package events

import "time"

type Actor struct {
	Kind  string `json:"kind"`
	ID    string `json:"id"`
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

type Subject struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
	Key  string `json:"key,omitempty"`
}

type Envelope struct {
	ID             string    `json:"id"`
	Type           string    `json:"type"`
	Version        string    `json:"version"`
	OccurredAt     time.Time `json:"occurred_at"`
	IdempotencyKey string    `json:"idempotency_key"`
	InstanceID     string    `json:"instance_id,omitempty"`
	Actor          *Actor    `json:"actor,omitempty"`
	Subject        *Subject  `json:"subject,omitempty"`
	Data           any       `json:"data"`
}
