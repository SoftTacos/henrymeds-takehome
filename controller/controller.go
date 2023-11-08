package controller

import (
	"context"
	"errors"
	"henrymeds-takehome/dao"
	"henrymeds-takehome/model"
	"log"
	"time"
)

const (
	reservationLeadTime       = time.Hour * 24
	reservationExpirationTime = time.Minute * 30
)

type Controller interface {
	CreateAvailabilities(ctx context.Context, request model.CreateAvailabilities) error
	GetAvailabilities(ctx context.Context, request model.GetAvailabilities) (availabilities []model.TimeRange, err error)
}

func NewController(dao dao.ReservationDao) *controller {
	return &controller{}
}

type controller struct {
	reservationDao dao.ReservationDao
}

func (c *controller) CreateAvailabilities(ctx context.Context, request model.CreateAvailabilities) (err error) {
	var (
		existingAvailabilities []model.TimeRange
		overlaps               []model.TimeRange
	)

	err = validateCreateAvailabilities(request)
	if err != nil {
		log.Println(err)
		return
	}

	// check if there are any overlapping availabilities
	existingAvailabilities, err = c.reservationDao.GetAvailabilities(ctx, model.GetAvailabilities{
		ProviderID: request.ProviderID,
		TimeRange:  request.TimeRange,
	})
	if err != nil {
		log.Println("failed to check for existing availabilities: ", err)
		return
	}

	overlaps = detectOverlap(append(existingAvailabilities, request.TimeRange))
	if len(overlaps) > 0 {
		// Ideally we don't throw an error here, we just extend the existing availability timerange to end when the submitted timerange ends
		// however, I only have 2h, so this corner is getting cut
		// TODO: give a better error here
		err = errors.New("requested availability overlaps with existing availability")
		log.Println(err)
		return
	}

	// insert
	err = c.reservationDao.InsertAvailabilities(ctx, request)
	// I prefer not to print every log on every layer unless it provides useful tracing context, this avoids log spam
	// the error will get logged on the handler layer
	return
}

func validateCreateAvailabilities(request model.CreateAvailabilities) (err error) {
	if !request.Start.Before(request.End) {
		err = errors.New("start time must be before end time")
	} else if request.Start.Minute()%15 != 0 {
		err = errors.New("start time minutes must be a multiple of 15")
	} else if request.End.Minute()%15 != 0 {
		err = errors.New("end time minutes must be a multiple of 15")
	} else if request.Start.Before(time.Now()) {
		err = errors.New("start time must be in the future")
	}

	return
}

func (c *controller) GetAvailabilities(ctx context.Context, request model.GetAvailabilities) (availabilities []model.TimeRange, err error) {

	err = validateGetAvailabilities(request)
	if err != nil {
		log.Println(err)
		return
	}

	// retrieve availabilities that overlap with request start-end
	availabilities, err = c.reservationDao.GetAvailabilities(ctx, request)
	if err != nil {
		log.Println("failed to retrieve availabilities: ", err)
		return
	}

	// retrieve reservations that overlap with request start-end

	// subtract reservations from retrieved availabilities

	return
}

func validateGetAvailabilities(request model.GetAvailabilities) (err error) {
	if !request.Start.Before(request.End) {
		err = errors.New("start time must be before end time")
	} else if request.Start.Minute()%15 != 0 {
		err = errors.New("start time minutes must be a multiple of 15")
	} else if request.End.Minute()%15 != 0 {
		err = errors.New("end time minutes must be a multiple of 15")
	}

	return
}

func (c *controller) CreateReservation(ctx context.Context, request model.CreateReservation) (confirmationID string, err error) {
	var (
		availabilities []model.TimeRange
		overlaps       []model.TimeRange
		// reservations   []model.Reservation
		newReservation model.Reservation
		confirmation   model.ReservationConfirmation
	)

	err = validateCreateReservation(request)
	if err != nil {
		log.Println(err)
		return
	}

	// retrieve availabilities that overlap with request timerange
	// error if reservation is not contained within availability
	availabilities, err = c.reservationDao.GetAvailabilities(ctx, model.GetAvailabilities{
		ProviderID: request.ProviderID,
		TimeRange:  request.TimeRange,
	})
	if err != nil {
		// TODO:error handling
		log.Println("failed to check for existing availabilities: ", err)
		return
	}

	overlaps = detectOverlap(append(availabilities, request.TimeRange))
	// if there is absolutely no overlap, that means there are no availabilities
	if len(overlaps) == 0 {
		err = errors.New("no availability during the requested time")
		log.Println(err)
		return
	} else {
		// figure out the total amount of overlap time
		var totalOverlapTime time.Duration
		for _, overlap := range overlaps {
			// IRL I would probably make a duration function for this
			totalOverlapTime += overlap.End.Sub(overlap.Start)
		}
		if totalOverlapTime != request.End.Sub(request.Start) {
			err = errors.New("insufficient availability during the requested time")
			log.Println(err)
			return
		}
	}

	// TODO: clean up the repeated calls

	err = c.checkReservationAvailability(ctx, request.ClientID, request.ProviderID, request.TimeRange)
	if err != nil {
		return
	}
	newReservation, err = c.reservationDao.InsertReservation(ctx, model.Reservation{
		ClientID:   request.ClientID,
		ProviderID: request.ProviderID,
		Confirmed:  false,
		TimeRange:  request.TimeRange,
	})
	if err != nil {
		log.Println("failed to insert reservation")
		return // TODO:error handling
	}

	confirmation, err = c.reservationDao.InsertReservationConfirmation(ctx, model.ReservationConfirmation{
		ReservationID: newReservation.ID,
		ExpiresAt:     time.Now().Add(reservationExpirationTime),
	})
	if err != nil {
		log.Println("failed to insert reservation confirmation")
		return // TODO:error handling
	}
	confirmationID = confirmation.ID

	return
}

func (c *controller) checkReservationAvailability(ctx context.Context, clientId string, providerId string, timerange model.TimeRange) (err error) {
	// TODO: check for collisions with self

	var reservations []model.Reservation
	reservations, err = c.reservationDao.GetReservations(ctx, model.GetReservations{
		ProviderID: providerId,
		TimeRange:  timerange,
	})
	if err != nil {
		log.Println("failed to retrieve reservations")
		return // TODO:error handling
	}
	// error if any are found
	if len(reservations) > 0 {
		err = errors.New("the provider has conflicting reservations during that time")
		log.Println(err)
		return
	}

	// retrieve reservations that overlap with request timerange
	reservations, err = c.reservationDao.GetReservations(ctx, model.GetReservations{
		ClientID:  clientId,
		TimeRange: timerange,
	})
	if err != nil {
		log.Println("failed to retrieve reservations")
		return // TODO:error handling
	}
	// error if any are found
	if len(reservations) > 0 {
		err = errors.New("the client has conflicting reservations during that time")
		log.Println(err)
	}
	return
}

func validateCreateReservation(request model.CreateReservation) (err error) {
	if !request.Start.Before(request.End) {
		err = errors.New("start time must be before end time")
	} else if request.Start.Minute()%15 != 0 {
		err = errors.New("start time minutes must be a multiple of 15")
	} else if request.End.Minute()%15 != 0 {
		err = errors.New("end time minutes must be a multiple of 15")
	} else if request.Start.Before(time.Now().Add(reservationLeadTime)) {
		err = errors.New("start time must be at least 24 hours before now")
	}

	return
}

func detectOverlap(timeRanges []model.TimeRange) (overlappingRanges []model.TimeRange) {
	var overlapStart, overlapEnd int64

	for i := 0; i < len(timeRanges)-1; i++ {
		for j := i + 1; j < len(timeRanges); j++ {
			if timeRanges[i].End.After(timeRanges[j].Start) && timeRanges[i].Start.Before(timeRanges[j].End) {
				// Time ranges overlap, add the overlapping range to the result
				overlapStart = max(timeRanges[i].Start.Unix(), timeRanges[j].Start.Unix())
				overlapEnd = min(timeRanges[i].End.Unix(), timeRanges[j].End.Unix())

				overlappingRanges = append(overlappingRanges, model.TimeRange{Start: time.Unix(overlapStart, 0), End: time.Unix(overlapEnd, 0)})
			}
		}
	}

	return
}

func (c *controller) ConfirmReservation(ctx context.Context, confirmationId string) (err error) {
	var (
		confirmation model.ReservationConfirmation
		reservations []model.Reservation
		reservation  model.Reservation
	)

	confirmation, err = c.reservationDao.GetReservationConfirmation(ctx, confirmationId)
	if err != nil {
		log.Println("failed to retrieve reservation confirmation")
		return // TODO:error handling
	}

	// get reservation
	reservations, err = c.reservationDao.GetReservations(ctx, model.GetReservations{
		ID: confirmation.ReservationID,
	})
	if err != nil {
		log.Println("failed to retrieve reservations")
		return // TODO:error handling
	}
	reservation = reservations[0]

	// if already confirmed, just return
	if reservation.Confirmed {
		log.Println("reservation already confirmed")
		return
	}

	// if it has NOT expired, confirm it and return
	if confirmation.ExpiresAt.Before(time.Now()) {
		err = c.reservationDao.ConfirmReservation(ctx, reservation.ID)
		if err != nil {
			log.Println("failed to confirm reservation: ", err)
		}
		return
	}

	// otherwise, it has expired, we need to check if other reservations have been booked in its place
	err = c.checkReservationAvailability(ctx, reservation.ClientID, reservation.ProviderID, reservation.TimeRange)
	if err != nil {
		return
	}

	err = c.reservationDao.ConfirmReservation(ctx, reservation.ID)
	if err != nil {
		log.Println("failed to confirm reservation: ", err)
	}

	return
}
