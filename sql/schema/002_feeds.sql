-- +goose Up
CREATE TABLE feeds (
    id UUID primary key,
    name TEXT not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    url TEXT not null unique,
    user_id UUID not null references users(id) on DELETE CASCADE
);


-- +goose Down
DROP TABLE users;