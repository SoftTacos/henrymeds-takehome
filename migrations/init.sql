-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4()
);

CREATE TABLE availabilities (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  provider_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  start TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  end TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE TABLE reservations (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  provider_id uuid NOT NULL REFERENCES users(id),
  client_id uuid NOT NULL REFERENCES users(id),
  start TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  end TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE TABLE confirmations (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  expires_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  reservation_id uuid NOT NULL REFERENCES reservations(id)
);

-- would put a trigger to verify that the time availabilities do not overlap

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE users;
DROP TABLE availabilities;