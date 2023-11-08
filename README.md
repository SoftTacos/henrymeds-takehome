# Dev Notes:
- I opted to skip any real use of users as a real concept. I left a users table in. However it isn't used anywhere. There are 4 default users created in the init.sql migration. The focus of this exercise was a reservation system so I stuck with that. This means that you can submit availiabilities for a client, which isn't a requested feature.
- In a real world setting you don't want to return IDs of any kind to the outside world. I demonstrated one method of doing so with how I handle confirmations. However due to the short timeframe I decided UUIDs would be good enough since a real-world implementation would have proper authentication and authorization, and we avoid the big problem of exposing sequential IDs.
- There are TODOs everywhere related to error handling. I'm tossing DB errors straight up the stack, which is not what I like to do. However handling errors-as-values takes a little more work, and there wasn't time
- There is poor code reuse in the handlers, that can be easily cleaned up with some util functions
- The general standard for configuration variables is to set them in environment variables. I decided to just pass them in as flags for clarity and ease of use. I don't want you to have to configure your environment just to run this once.
- In a real-world setting, I would have changed "Availability" to "Opening" or something shorter. I decided to keep it for this exercise for continuity.
- I defintely went over on dev-time, I took appx 2h 40min on the code. I felt that compromising on dev time was an acceptable requirements tradeoff vs no catching edge cases.
- Upon writing this README, I realized I misinterpreted one of the requirements in my haste. Instead of all reservations being 15min long, I enforced that all reservations and availiabilities to start and end on 15min intervals. This would be a pretty simple fix, but I want to stick to my 2:40 time estimate.
- Retrieval of availabilities does not subtract the reservations from available times. I intended to circle back and implement that, however I did not notice that until now.

## Setup Notes
I set this whole thing up on Linux, if you are on windows you might have to adapt some of these setup steps.

Prerequisites:
- postgresql
- postgres user with CREATEDB permissions


## Setup: 
- Install Golang: https://go.dev/doc/install
- I'm using golang's Goose to manage migrations. Since there's only one migration, you could just create and copy paste the migration into the DB. I won't tell. Just run `go get github.com/pressly/goose && go install github.com/pressly/goose`
- > make setup

## Run:
- > cd henrymeds-takehome
- > go run . -port 9001 -db `db_url`

Connection string format: postgresql://[user[:password]@][netloc][:port][/dbname][?param1=value1&...]

Example db_url: postgres://postgres:postgres_password@localhost/henrymed?sslmode=disable

# Endpoints:
- variables are highlighted or surrounded by backticks, depending on if you're using a .md viewer
- all times are in RFC3339 format. Ex: `2023-11-11T15:15:00Z`

## Get availabilities
Format: GET /users/`providerId`/availabilities?start=`start_time`&end=`end_time`

Example URL: http://localhost:9001/users/e1ceaf4f-b5a5-4848-a71b-82b2ef02dd5e/availabilities?start=2023-11-10T15:15:00Z&end=2023-11-11T15:15:00Z

Example Response Body:
```
[
    {
        "start": "2023-11-11T15:15:00Z",
        "end": "2023-11-12T15:15:00Z"
    },
    {
        "start": "2023-11-10T15:15:00Z",
        "end": "2023-11-11T15:15:00Z"
    }
]
```


## Create availabilities
Format: POST /users/`providerId`/availabilities
Body: 
```
{
    "start":"2023-11-11T15:15:00Z",
    "end":"2023-11-12T15:15:00Z"
}
```

Response body is empty

Example URL: http://localhost:9001/users/e1ceaf4f-b5a5-4848-a71b-82b2ef02dd5e/availabilities

## Create reservation
Format: POST /reservations
Body: 
```
{
    "clientId":"aa5ad430-a5f5-4a80-ad84-f22bc2852966",
    "providerId":"e1ceaf4f-b5a5-4848-a71b-82b2ef02dd5e",
    "start":"2023-11-11T15:15:00Z",
    "end":"2023-11-12T15:15:00Z"
}
```

Returns the confirmationId in the response body as a string

Example Response Body:
```
66fb346e-fb17-41b6-8cff-fe9d3ae104f4
```

## Create availabilities
Format: POST /reservations/confirm/`confirmationId`
Body: 
```
{
    "start":"2023-11-11T15:15:00Z",
    "end":"2023-11-12T15:15:00Z"
}
```

Response body is empty

Example URL: http://localhost:9001/reservations/confirm/66fb346e-fb17-41b6-8cff-fe9d3ae104f4
