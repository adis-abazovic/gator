-- +goose Up
CREATE TABLE posts(
    id uuid PRIMARY KEY NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    published_at TIMESTAMP,
    title TEXT,
    url TEXT UNIQUE NOT NULL,
    description TEXT,
    feed_id UUID NOT NULL,
    FOREIGN KEY (feed_id)
    REFERENCES feeds(id)
    ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;