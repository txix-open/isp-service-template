-- +goose Up

CREATE TABLE objects
(
    id      serial8 PRIMARY KEY ,
    name    varchar(64)
);

-- +goose Down

DROP TABLE objects CASCADE;