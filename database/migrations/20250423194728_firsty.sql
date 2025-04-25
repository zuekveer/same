-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied.
CREATE TABLE users (
                       id UUID PRIMARY KEY,
                       name TEXT NOT NULL,
                       age INT NOT NULL
);

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back.
DROP TABLE users;