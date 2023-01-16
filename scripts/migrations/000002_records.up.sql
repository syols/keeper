CREATE TYPE detail AS ENUM ('TEXT', 'BLOB', 'CARD', 'DETAIL');

CREATE TABLE records (
                         id SERIAL PRIMARY KEY,
                         user_id INT REFERENCES users (id),
                         metadata TEXT NOT NULL UNIQUE,
                         detail detail NOT NULL
);

CREATE TABLE text_details (
                              id SERIAL PRIMARY KEY,
                              record_id INT REFERENCES records (id) UNIQUE,
                              data TEXT
);

CREATE TABLE blob_details (
                              id SERIAL PRIMARY KEY,
                              record_id INT REFERENCES records (id) UNIQUE,
                              data BYTEA
);

CREATE TABLE card_details (
                              id SERIAL PRIMARY KEY,
                              record_id INT REFERENCES records (id) UNIQUE,
                              number TEXT,
                              cardholder TEXT,
                              cvc int,
                              expiration TIMESTAMP WITH TIME ZONE
);

CREATE TABLE login_details (
                               id SERIAL PRIMARY KEY,
                               record_id INT REFERENCES records (id) UNIQUE,
                               login TEXT,
                               password TEXT
);
