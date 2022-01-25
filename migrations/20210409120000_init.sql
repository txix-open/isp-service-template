-- +goose Up
CREATE TABLE object
(
    id   SERIAL8 PRIMARY KEY,
    name TEXT
);

-- +goose Down
DROP TABLE object;
