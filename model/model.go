package model

import "time"

type User struct {
	ID string
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}

type Reservation struct {
	ID         string
	ClientID   string
	ProviderID string
	Confirmed  bool
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
	ReservationID string
	ClientID      string
	ProviderID    string
	TimeRange
}

type GetReservations struct {
	ID         string
	ProviderID string
	ClientID   string
	TimeRange
}

type Availability struct {
	ID string
	TimeRange
}

type ReservationConfirmation struct {
	ID            string
	ReservationID string
	ExpiresAt     time.Time
}
