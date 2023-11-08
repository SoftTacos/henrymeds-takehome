-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  username VARCHAR(50) NOT NULL
);

CREATE TABLE availabilities (
  -- don't need an ID here but will leave it in, could opt for a unique (provider_id, start, end) constraint since nothing will be null
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  provider_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  start_time TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  end_time TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  -- these should be enforced in code, however sometimes people like to mess with the DB directly
  UNIQUE(provider_id,start_time,end_time)
  -- could put a trigger to verify that the time availabilities do not overlap
);
CREATE INDEX availabilities_times ON availabilities (start_time,end_time);

CREATE TABLE reservations (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  confirmation_id uuid DEFAULT uuid_generate_v4(),
  expires_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  provider_id uuid NOT NULL REFERENCES users(id),
  confirmed boolean NOT NULL DEFAULT false,
  client_id uuid NOT NULL REFERENCES users(id),
  start_time TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  end_time TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  -- these should be enforced in code, however sometimes people like to mess with the DB directly
  UNIQUE(provider_id,start_time,end_time),
  UNIQUE(client_id,start_time,end_time),
  UNIQUE(confirmation_id)
);
CREATE INDEX reservations_times ON reservations (start_time,end_time);

-- setup some starter users
INSERT INTO users (id, username) VALUES
  ('e1ceaf4f-b5a5-4848-a71b-82b2ef02dd5e','provider1'),
  ('f4bc7e96-6a6b-4872-ba07-207b49a95444','provider2'),
  ('aa5ad430-a5f5-4a80-ad84-f22bc2852966','client1'),
  ('067de952-733b-4113-9542-5bc26133722c','client2');

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE users;
DROP TABLE availabilities;
DROP TABLE reservation_confirmations;
DROP TABLE reservations;