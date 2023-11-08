package controller

import (
	"context"
	"errors"
	"fmt"
	"henrymeds-takehome/dao"
	"henrymeds-takehome/model"
	"log"
	"time"

	"github.com/google/uuid"
)

const (
	reservationLeadTime       = time.Hour * 24
	reservationExpirationTime = time.Minute * 30
)

type Controller interface {
	CreateAvailability(ctx context.Context, request model.CreateAvailabilities) error
	GetAvailabilities(ctx context.Context, request model.GetAvailabilities) (availabilities []model.TimeRange, err error)
	CreateReservation(ctx context.Context, request model.CreateReservation) (confirmationID string, err error)
	ConfirmReservation(ctx context.Context, confirmationId string) (err error)
}

func NewController(dao dao.ReservationDao) *controller {
	return &controller{
		reservationDao: dao,
	}
}

type controller struct {
	reservationDao dao.ReservationDao
}

func (c *controller) CreateAvailability(ctx context.Context, request model.CreateAvailabilities) (err error) {
	var (
		existingAvailabilities []model.Availability
		overlaps               []model.TimeRange
	)

	err = validateCreateAvailability(request)
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
		return
	}

	overlaps = detectOverlap(append(extractTimeranges(existingAvailabilities), request.TimeRange))
	if len(overlaps) > 0 {
		// Ideally we don't throw an error here, we just extend the existing availability timerange to end when the submitted timerange ends
		// however, I only have 2h, so this corner is getting cut
		// TODO: give a better error here
		err = errors.New("requested availability overlaps with existing availability")
		log.Println(err)
		return
	}

	// insert
	err = c.reservationDao.InsertAvailabilities(ctx, []model.Availability{
		{
			TimeRange:  request.TimeRange,
			ProviderID: request.ProviderID,
		},
	})
	// I prefer not to print every log on every layer unless it provides useful tracing context, this avoids log spam
	// the error will get logged on the handler layer
	return
}

func validateCreateAvailability(request model.CreateAvailabilities) (err error) {
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

func (c *controller) GetAvailabilities(ctx context.Context, request model.GetAvailabilities) (availableTimeranges []model.TimeRange, err error) {
	var (
		existingAvailabilities []model.Availability
	)
	err = validateGetAvailabilities(request)
	if err != nil {
		log.Println(err)
		return
	}

	// retrieve availabilities that overlap with request start-end
	existingAvailabilities, err = c.reservationDao.GetAvailabilities(ctx, request)
	if err != nil {
		log.Println("failed to retrieve availabilities: ", err)
		return
	}
	availableTimeranges = extractTimeranges(existingAvailabilities)
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
		availabilities []model.Availability
		overlaps       []model.TimeRange
		// reservations   []model.Reservation
		newReservation model.Reservation
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

	// detect overlaps and verify that the reservation is completely contained within availabilities
	overlaps = detectOverlap(append(extractTimeranges(availabilities), request.TimeRange))
	// if there is absolutely no overlap, that means there are no availabilities
	if len(overlaps) == 0 {
		err = fmt.Errorf("no availability during the requested times: %v - %v", request.Start, request.End)
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

	err = c.checkReservationAvailability(ctx, request.ClientID, request.ProviderID, request.TimeRange)
	if err != nil {
		return
	}
	newReservation, err = c.reservationDao.InsertReservation(ctx, model.Reservation{
		ClientID:   request.ClientID,
		ProviderID: request.ProviderID,
		ExpiresAt:  time.Now().Add(reservationExpirationTime),
		Confirmed:  false,
		TimeRange:  request.TimeRange,
	})
	if err != nil {
		log.Println("failed to insert reservation")
		return // TODO:error handling
	}

	confirmationID = newReservation.ConfirmationID

	return
}

func (c *controller) checkReservationAvailability(ctx context.Context, clientId string, providerId string, timerange model.TimeRange) (err error) {
	// TODO: check for collisions with self

	var reservations []model.Reservation
	reservations, err = c.reservationDao.GetReservations(ctx, model.GetReservations{
		ProviderID: providerId,
		TimeRange:  &timerange,
	})
	if err != nil {
		log.Println("failed to retrieve reservations")
		return // TODO:error handling
	}
	// if there are any reservations that have been confirmed or haven't expired
	// personally I would prefer to put this in the query, but that complicates the function
	// significantly and adds significant devtime for this exercise
	if len(reservations) > 0 {
		for _, res := range reservations {
			if res.Confirmed || time.Now().After(res.ExpiresAt) {
				err = errors.New("the provider has conflicting reservations during that time")
				log.Println(err)
				return
			}
		}
	}

	// retrieve reservations that overlap with request timerange
	reservations, err = c.reservationDao.GetReservations(ctx, model.GetReservations{
		ClientID:  clientId,
		TimeRange: &timerange,
	})
	if err != nil {
		log.Println("failed to retrieve reservations")
		return // TODO:error handling
	}
	// error if more than 1 are found, don't allow a client to double book even unconfirmed reservations.
	if len(reservations) > 1 {
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

// I separated confirmations out into ReservationConfirmation to demonstrate some level of security competency
// I realize I use UUIDs everywhere else, however generally you want to try to hide as many ID/UUIDs as possible from the
// outside world. There wasn't time to do that however, and the user only uses their own UUID and their provider's UUID.
func (c *controller) ConfirmReservation(ctx context.Context, confirmationId string) (err error) {
	var (
		reservations []model.Reservation
		reservation  model.Reservation
	)

	// validate the UUID, don't want strings going directly to the DB
	_, err = uuid.Parse(confirmationId)
	if err != nil {
		log.Println("invalid UUID provided")
		err = errors.New("invalid UUID provided")
		return
	}

	// get reservation
	reservations, err = c.reservationDao.GetReservations(ctx, model.GetReservations{
		ConfirmationID: confirmationId,
	})
	if err != nil {
		log.Println("failed to retrieve reservations")
		return // TODO:error handling
	}
	if len(reservations) == 0 {
		err = errors.New("no reservation found for that confirmation Id")
		log.Println(err.Error(), confirmationId)
		return
	}
	reservation = reservations[0]

	// if already confirmed, just return
	if reservation.Confirmed {
		log.Println("reservation already confirmed")
		return
	}

	// if it has NOT expired, confirm it and return
	if reservation.ExpiresAt.Before(time.Now()) {
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

func extractTimeranges(availabilities []model.Availability) (timeranges []model.TimeRange) {
	for _, avail := range availabilities {
		timeranges = append(timeranges, avail.TimeRange)
	}
	return
}
