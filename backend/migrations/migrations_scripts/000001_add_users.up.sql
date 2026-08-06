CREATE TABLE IF NOT EXISTS users (
    user_id serial primary key,
    username VARCHAR(40) not null UNIQUE ,
    password_hash VARCHAR(200) not null,
    created_at TIMESTAMP
);