-- +goose Up
CREATE TABLE message
(
    id      int8 primary key,
    version int8  not null,
    data    jsonb not null
);

-- +goose Down
DROP TABLE message;
