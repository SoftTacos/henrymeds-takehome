# Dev Notes:
- I opted to skip any real use of users as a real concept. I left a users table in. However it isn't used anywhere. There are 4 default users created in the init.sql migration. The focus of this exercise was a reservation system so I stuck with that. This means that you can submit availiabilities for a client, which isn't a requested feature.
- In a real world setting you don't want to return IDs of any kind to the outside world. I demonstrated one method of doing so with how I handle confirmations. However due to the short timeframe I decided UUIDs would be good enough since a real-world implementation would have proper authentication and authorization, and we avoid the big problem of exposing sequential IDs.
- There are TODOs everywhere related to error handling. I'm tossing DB errors straight up the stack, which is not what I like to do. However handling errors-as-values takes a little more work, and there wasn't time
- There is poor code reuse in the handlers, that can be easily cleaned up with some util functions
- The general standard for configuration variables is to set them in environment variables. I decided to just pass them in as flags for clarity and ease of use. I don't want you to have to configure your environment just to run this once.
- In a real-world setting, I would have changed "Availability" to "Opening" or something shorter. I decided to keep it for this exercise for continuity.
- I defintely went over on dev-time, I took appx 2h 40min on the code. I felt that compromising on dev time was an acceptable requirements tradeoff vs no catching edge cases.

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

## Endpoints:
- TODO

