-- Migration name: custom_migration_name

CREATE TABLE my_table
(
    id        UUID PRIMARY KEY,
    something TEXT NOT NULL
);
INSERT INTO my_table (id, something)
VALUES (gen_random_uuid(), 'hello WORLD'),
       (gen_random_uuid(), 'hello another world')