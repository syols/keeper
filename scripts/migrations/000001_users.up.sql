CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       login TEXT NOT NULL UNIQUE,
                       password TEXT NOT NULL
);
