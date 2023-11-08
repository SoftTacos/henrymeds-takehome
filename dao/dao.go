package dao

import (
	"context"
	"henrymeds-takehome/model"
)

type ReservationDao interface {
	InsertAvailabilities(ctx context.Context, request model.CreateAvailabilities) error
	GetAvailabilities(ctx context.Context, request model.GetAvailabilities) (availabilities []model.TimeRange, err error)
	InsertReservation(ctx context.Context, reservation model.Reservation) (model.Reservation, error)
	GetReservations(ctx context.Context, request model.GetReservations) (reservations []model.Reservation, err error)
	//todo
	ConfirmReservation(ctx context.Context, reservationId string) error
	InsertReservationConfirmation(ctx context.Context, confirmation model.ReservationConfirmation) (model.ReservationConfirmation, error)
	GetReservationConfirmation(ctx context.Context, confirmationId string) (model.ReservationConfirmation, error)
	GetUsers(ctx context.Context, id string) (user model.User, err error)
}

func NewReservationDao() *reservationDao {
	return &reservationDao{}
}

type reservationDao struct {
}
