package handler

import (
	"fmt"
	"henrymeds-takehome/controller"
	"henrymeds-takehome/model"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func NewHandler(controller controller.Controller) *Handler {
	return &Handler{
		controller: controller,
	}
}

const (
	ProviderIdParam = "providerId"
	StartParam      = "start"
	EndParam        = "end"

	errInvalidTimeFormat = "invalid time format provided, please use RFC3339"
)

// this really should be a reservation handler and an availability handler
type Handler struct {
	controller controller.Controller
}

func (h *Handler) HandleGetAvailabilitiesRequest(c echo.Context) (err error) {
	var (
		availabilities []model.TimeRange
		times          []time.Time
	)
	// parse the JSON into a timerange
	times, err = parseTimes([]string{
		c.QueryParam(StartParam),
		c.QueryParam(EndParam),
	})
	if err != nil {
		_ = c.String(http.StatusBadRequest, errInvalidTimeFormat)
		return
	}

	availabilities, err = h.controller.GetAvailabilities(c.Request().Context(), model.GetAvailabilities{
		TimeRange: model.TimeRange{
			Start: times[0],
			End:   times[1],
		},
		ProviderID: c.Param(ProviderIdParam),
	})
	if err != nil {
		// TODO: status handler
		msg := fmt.Sprintf("failed to get availabilities: %s", err.Error())
		_ = c.String(http.StatusBadRequest, msg)
		return
	}
	_ = c.JSON(http.StatusOK, availabilities)
	return
}

func (h *Handler) HandleCreateAvailabilityRequest(c echo.Context) (err error) {
	var (
		request = model.TimeRange{}
	)
	err = c.Bind(&request)
	if err != nil {
		msg := fmt.Sprintf("failed to parse create availabilities request: %s", err.Error())
		log.Println(msg)
		_ = c.String(http.StatusBadRequest, msg)
		return
	}
	fmt.Println(c.Param(ProviderIdParam))
	err = h.controller.CreateAvailability(c.Request().Context(), model.CreateAvailabilities{
		TimeRange:  request,
		ProviderID: c.Param(ProviderIdParam),
	})
	if err != nil {
		// TODO: status handler
		msg := fmt.Sprintf("failed to create availability: %s", err.Error())
		_ = c.String(http.StatusBadRequest, msg)
		return
	}
	_ = c.NoContent(http.StatusOK)
	return
}

func (h *Handler) HandleCreateReservationRequest(c echo.Context) (err error) {
	var (
		request        = model.CreateReservation{}
		confirmationID string
	)
	err = c.Bind(&request)
	if err != nil {
		msg := fmt.Sprintf("failed to parse create reservations request: %s", err.Error())
		log.Println(msg)
		_ = c.String(http.StatusBadRequest, msg)
		return
	}

	confirmationID, err = h.controller.CreateReservation(c.Request().Context(), request)
	if err != nil {
		// TODO: status handler
		msg := fmt.Sprintf("failed to create reservation: %s", err.Error())
		_ = c.String(http.StatusBadRequest, msg)
		return
	}
	_ = c.String(http.StatusOK, confirmationID)
	return
}

func (h *Handler) HandleConfirmReservationRequest(c echo.Context) (err error) {
	var confirmationId string

	confirmationId = c.Param("confirmationId")
	if err != nil {
		msg := fmt.Sprintf("failed to parse create reservations request: %s", err.Error())
		log.Println(msg)
		_ = c.String(http.StatusBadRequest, msg)
		return
	}
	err = h.controller.ConfirmReservation(c.Request().Context(), confirmationId)
	if err != nil {
		// TODO: status handler
		msg := fmt.Sprintf("failed to confirm reservation: %s", err.Error())
		_ = c.String(http.StatusBadRequest, msg)
		return
	}
	_ = c.NoContent(http.StatusOK)
	return
}

// everything below here would go into a util package

// parses list of times
func parseTimes(timeStrs []string) (times []time.Time, err error) {
	for _, ts := range timeStrs {
		var t time.Time
		t, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			log.Println("invalid time provided: ", err)
			return
		}
		times = append(times, t)
	}

	return
}
