-- type Demand struct {
-- 	Slots        []Slot    `json:"slots"`
-- 	VkID         int       `json:"vk_id"`
-- 	FirstName    string    `json:"first_name"`
-- 	LastName     string    `json:"last_name"`
-- 	ForWeek      time.Time `json:"for_week"`
-- 	PlayersCount int       `json:"players_count"`
-- }


CREATE TABLE IF NOT EXISTS demands (
    id serial,
    vk_id int not null,
    for_week TIMESTAMP not null,
    first_name VARCHAR(40),
    last_name VARCHAR(40),
    players_count int not null,
    slots jsonb,

    PRIMARY KEY (id, vk_id,for_week )
);


-- type Slot struct {
-- 	ID         int    `json:"id"`
-- 	Name       string `json:"name"`
-- 	ValidFrom  string `json:"valid_from"`
-- 	ValidUntil string `json:"valid_until"`
-- }

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


-- type Master struct {
-- 	Name string `json:"name"`
-- 	VkID string `json:"vk_id"`
-- }

CREATE TABLE IF NOT EXISTS masters (
    id serial  ,
    name VARCHAR(40) not null,
    vk_id int not null primary key,
    created_at timestamp not null,
    updated_at timestamp not null
);