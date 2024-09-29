CREATE TABLE IF NOT EXISTS users (
     email varchar PRIMARY KEY,
     password_hash varchar,
     created_at timestamp NOT NULL,
     updated_at timestamp NOT NULL DEFAULT (timezone('utc', now())),
    urls_left int NOT NULL DEFAULT 10

    );

CREATE TABLE IF NOT EXISTS urls (
    short_url varchar PRIMARY KEY,
    long_url varchar NOT NULL,
    user_email varchar NOT NULL REFERENCES users(email),
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL DEFAULT (timezone('utc', now())),
    expires_at timestamp NOT NULL
);


CREATE TABLE IF NOT EXISTS subscriptions (
    id int PRIMARY KEY,
    name varchar NOT NULL,
    total_urls int NOT NULL
);