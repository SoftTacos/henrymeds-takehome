package dao

import (
	"context"
	"fmt"
	"henrymeds-takehome/model"
	"log"

	gopg "github.com/go-pg/pg/v10"
)

type ReservationDao interface {
	InsertAvailabilities(context.Context, []model.Availability) error
	GetAvailabilities(context.Context, model.GetAvailabilities) ([]model.Availability, error)
	InsertReservation(context.Context, model.Reservation) (model.Reservation, error)
	GetReservations(context.Context, model.GetReservations) ([]model.Reservation, error)
	GetUser(ctx context.Context, id string) (user model.User, err error)
	// UpdateReservation would be a more generalized way to do this, but I took a shortcut
	ConfirmReservation(ctx context.Context, reservationId string) error
}

func NewReservationDao(db *gopg.DB) *dao {
	return &dao{
		db: db,
	}
}

type dao struct {
	db *gopg.DB
}

func (d *dao) InsertAvailabilities(ctx context.Context, request []model.Availability) (err error) {
	_, err = d.db.Model(&request).Insert()
	if err != nil {
		log.Println("failed to insert availabilities: ", err)
	}
	return
}

func (d *dao) GetAvailabilities(ctx context.Context, request model.GetAvailabilities) (availabilities []model.Availability, err error) {

	// SELECT * FROM availabilities WHERE
	// [ provider_id = ? | (?,?) OVERLAPS (start,end) ]

	var query = d.db.Model(&availabilities)
	// if provider ID was set, add it to the query
	if request.ProviderID != "" {
		query.Where("provider_id = ?", request.ProviderID)
	}
	if request.TimeRange.Start.Unix() != 0 {
		query.Where("(?,?) OVERLAPS (start_time,end_time)", request.Start, request.End)
	}

	err = query.Select()
	if err != nil {
		log.Println("failed to retrieve availabilities: ", err)
	}
	return
}

func (d *dao) InsertReservation(ctx context.Context, reservation model.Reservation) (model.Reservation, error) {
	var err error

	_, err = d.db.Model(&reservation).Insert()
	if err != nil {
		log.Println("failed to insert reservation: ", err)
	}
	return reservation, err
}

func (d *dao) GetReservations(ctx context.Context, request model.GetReservations) (reservations []model.Reservation, err error) {
	var query = d.db.Model(&reservations)

	// if ID was set, add it to the query
	if request.ID != "" {
		query.Where("id = ?", request.ID)
	}
	if request.ProviderID != "" {
		query.Where("provider_id = ?", request.ProviderID)
	}
	if request.ClientID != "" {
		query.Where("client_id = ?", request.ClientID)
	}
	if request.TimeRange != nil {
		fmt.Println(request.TimeRange.Start)
		query.Where("(?,?) OVERLAPS (start_time,end_time)", request.Start, request.End)
	}

	err = query.Select()
	return
}

func (d *dao) GetUser(ctx context.Context, id string) (user model.User, err error) {
	err = d.db.Model(&user).Where("id = ?", id).Select()
	if err != nil {
		log.Println("failed to retrieve user: ", err)
	}
	return
}

func (d *dao) ConfirmReservation(ctx context.Context, reservationId string) (err error) {
	_, err = d.db.Model(&model.Reservation{}).Where("id = ?", reservationId).Set("confirmed = ?", true).Update()
	if err != nil {
		log.Println("failed to confirm reservation: ", err)
	}
	return
}
