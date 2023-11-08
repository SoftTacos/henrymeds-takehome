package main

import (
	"flag"
	"fmt"
	c "henrymeds-takehome/controller"
	d "henrymeds-takehome/dao"
	h "henrymeds-takehome/handler"
	"log"
	"time"

	gopg "github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
)

var (
	port  = flag.String("port", "9001", "the port that the service will listen on")
	dbUrl = flag.String("db", "", "the database connection string")
)

func main() {
	config := readConfigs()
	handler := setupService(config)
	e := setupServer(handler)
	// start the server, e.Start() returns an error
	// e.Start() pauses the main goroutine(like a thread)
	// and runs the server in a separate goroutine
	// normally I'd have a sig term litener setup to shutdown
	// gracefully but I'm out of time
	log.Println(e.Start(":" + config.port))
}

type config struct {
	port  string
	dbUrl string
}

func readConfigs() config {
	flag.Parse()
	if *dbUrl == "" {
		panic("db flag not set, please provide a db url, see README for more info")
	}
	if *port == "" {
		fmt.Println("port not set, defaulting to 9001")
	}
	return config{
		port:  *port,
		dbUrl: *dbUrl,
	}
}

func setupService(config config) *h.Handler {
	db, err := createGoPgDB(config.dbUrl)
	if err != nil {
		panic("failed to setup DB connection:" + err.Error())
	}
	dao := d.NewReservationDao(db)
	controller := c.NewController(dao)
	handler := h.NewHandler(controller)
	return handler
}

func setupServer(handler *h.Handler) (e *echo.Echo) {
	e = echo.New()
	e.Router().Add("GET", "/users/:providerId/availabilities", handler.HandleGetAvailabilitiesRequest)
	e.Router().Add("POST", "/users/:providerId/availabilities", handler.HandleCreateAvailabilityRequest)
	e.Router().Add("POST", "/reservations", handler.HandleCreateReservationRequest)
	e.Router().Add("POST", "/reservations/confirm/:confirmationId", handler.HandleConfirmReservationRequest)
	return
}

func createGoPgDB(url string) (*gopg.DB, error) {
	options, err := gopg.ParseURL(url)
	if err != nil {
		return nil, err
	}
	options.DialTimeout = 20 * time.Second
	db := gopg.Connect(options)

	// check connection
	_, err = db.Exec("SELECT 1")
	if err != nil {
		return nil, err
	}

	return db, nil
}
