package model

import "time"

type User struct {
	ID       string
	Username string
}

type TimeRange struct {
	Start time.Time `json:"start" pg:"start_time"`
	End   time.Time `json:"end" pg:"end_time"`
}

type Reservation struct {
	ID             string
	ClientID       string
	ProviderID     string
	Confirmed      bool
	ConfirmationID string
	ExpiresAt      time.Time
	TimeRange
}

type CreateAvailabilities struct {
	ProviderID string
	TimeRange
}

type GetAvailabilities struct {
	ProviderID string
	TimeRange
}

type CreateReservation struct {
	ClientID   string `json:"clientId"`
	ProviderID string `json:"providerId"`
	TimeRange
}

type GetReservations struct {
	ID             string
	ProviderID     string
	ClientID       string
	ConfirmationID string
	*TimeRange
}

type Availability struct {
	ID         string
	ProviderID string
	TimeRange
}
