package handler

import (
	"fmt"
	"henrymeds-takehome/controller"
	"henrymeds-takehome/model"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// not interfaced because we don't need to mock
type Handler struct {
	controller controller.Controller
}

func (h *Handler) HandleGetAvailabilitiesRequest(c echo.Context) (err error) {
	var (
		request        = model.GetAvailabilities{}
		availabilities []model.TimeRange
	)
	err = c.Bind(&request)
	if err != nil {
		msg := fmt.Sprintf("failed to parse get availabilities request: %s", err.Error())
		log.Println(msg)
		_ = c.String(http.StatusBadRequest, msg)
		return
	}

	availabilities, err = h.controller.GetAvailabilities(c.Request().Context(), request)
	if err != nil {
		// TODO: status handler
		msg := fmt.Sprintf("failed to get availabilities: %s", err.Error())
		log.Println(msg)
		_ = c.String(http.StatusBadRequest, msg)
		return
	}
	_ = c.JSON(http.StatusOK, availabilities)
	return
}

func (h *Handler) HandleCreateAvailabilitiesRequest(c echo.Context) (err error) {
	var (
		request = model.CreateAvailabilities{}
	)
	err = c.Bind(&request)
	if err != nil {
		msg := fmt.Sprintf("failed to parse create availabilities request: %s", err.Error())
		log.Println(msg)
		_ = c.String(http.StatusBadRequest, msg)
		return
	}

	err = h.controller.CreateAvailabilities(c.Request().Context(), request)
	if err != nil {
		// TODO: status handler
		msg := fmt.Sprintf("failed to get availabilities: %s", err.Error())
		log.Println(msg)
		_ = c.String(http.StatusBadRequest, msg)
		return
	}
	_ = c.NoContent(http.StatusOK)
	return
}
