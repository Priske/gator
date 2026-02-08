-- +goose Up
CREATE TABLE posts (
    id UUID primary key,
    created_at timestamp not null,
    updated_at timestamp not null,
    title TEXT  not null,
    url TEXT unique not null,
    description TEXT ,
    published_at timestamp ,
    feed_id UUID not null references feeds(id) on DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;