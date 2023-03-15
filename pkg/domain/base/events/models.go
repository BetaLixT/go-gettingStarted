// Package events contains the base event model
package events

import "time"

// EventEntity base event model
type EventEntity struct {
	Id        uint64    `json:"id"`
	SagaId    *string   `json:"sagaId"`
	Stream    string    `json:"stream"`
	StreamId  string    `json:"streamId"`
	Event     string    `json:"event"`
	Version   uint64    `json:"version"`
	EventTime time.Time `json:"eventTime"`
}
