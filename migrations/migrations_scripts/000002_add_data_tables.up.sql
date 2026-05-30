CREATE TABLE IF NOT EXISTS demands (
    id serial,
    vk_id int not null,
    for_week TIMESTAMP not null,
    first_name VARCHAR(40),
    last_name VARCHAR(40),
    players_count int not null,
    slots jsonb,
    created_at timestamp not null,
    updated_at timestamp not null,

    PRIMARY KEY (vk_id,for_week )
);

CREATE TABLE IF NOT EXISTS slots (
    id int not null,
    name VARCHAR(40) not null,
    valid_from TIMESTAMP not null,
    valid_until TIMESTAMP not null,

    PRIMARY KEY (id, valid_from )
);


CREATE TABLE IF NOT EXISTS slots_default (
    id int not null primary key ,
    name VARCHAR(40) not null
);

CREATE TABLE IF NOT EXISTS masters (
    id serial  ,
    name VARCHAR(40) not null,
    vk_id int not null primary key,
    created_at timestamp not null,
    updated_at timestamp not null
);