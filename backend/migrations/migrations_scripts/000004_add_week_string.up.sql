BEGIN;

ALTER TABLE demands ADD COLUMN for_week_str varchar(50) NOT NULL DEFAULT '';

ALTER TABLE demands
    DROP CONSTRAINT demands_pkey;

ALTER TABLE demands
    ADD CONSTRAINT demands_pkey PRIMARY KEY (vk_id, for_week_str);

COMMIT;