-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE USERS (
   id uuid DEFAULT uuid_generate_v4 (),
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL,
   name VARCHAR NOT NULL
);

-- +goose Down
DROP TABLE USERS;
