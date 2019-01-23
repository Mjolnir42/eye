-- SCHEMA VERSION: 201607010001
--
-- connect as RDBMS superuser
--
-- create roles for running eye
\connect postgres
CREATE ROLE eye_dba WITH NOSUPERUSER NOCREATEDB NOCREATEROLE LOGIN ENCRYPTED PASSWORD 'veryStrongAndSecretPassword';
CREATE ROLE eye_service WITH NOSUPERUSER NOCREATEDB NOCREATEROLE LOGIN ENCRYPTED PASSWORD 'similarlyStrongAndSecretPassword';
--
-- create database
CREATE DATABASE eye WITH OWNER eye_dba ENCODING 'UTF8' LC_COLLATE 'en_US.UTF-8' LC_CTYPE 'en_US.UTF-8' TEMPLATE template0;
GRANT CONNECT ON DATABASE eye TO eye_dba;
GRANT CONNECT ON DATABASE eye TO eye_service;
--
-- reconnect as eye_dba user (DB Owner)
\connect eye
--
-- setup schema eye
CREATE SCHEMA IF NOT EXISTS eye;
SET search_path TO eye;
ALTER DATABASE eye SET search_path TO eye;
--
-- create table configuration_lookup
CREATE TABLE IF NOT EXISTS eye.configuration_lookup (
  lookup_id               char(64)        PRIMARY KEY,
  host_id                 numeric(16,0)   NOT NULL,
  metric                  text            NOT NULL
);
--
-- create table configuration_items
CREATE TABLE IF NOT EXISTS eye.configuration_items (
  configuration_item_id   uuid            PRIMARY KEY,
  lookup_id               char(64)        NOT NULL REFERENCES eye.configuration_lookup( lookup_id ),
  configuration           jsonb           NOT NULL
);
--
-- create lookup acceleration index
CREATE INDEX _item_lookup ON eye.configuration_items (
  lookup_id,
  configuration_item_id
);
--
-- create schema version registry
CREATE TABLE IF NOT EXISTS public.schema_versions (
  serial                  bigserial       PRIMARY KEY,
  schema                  varchar(16)     NOT NULL,
  version                 numeric(16,0)   NOT NULL,
  created_at              timestamptz(3)  NOT NULL DEFAULT NOW()::timestamptz(3),
  description             text            NOT NULL
);
--
-- register schema version installation
INSERT INTO public.schema_versions (
  schema,
  version,
  description
) VALUES (
  'eye',
  201607010001,
  'Initial setup via: db-schema.201607010001.sql'
);
--
-- allow service account to use the database
GRANT INSERT, SELECT, UPDATE, DELETE ON ALL TABLES IN SCHEMA eye TO eye_service;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO eye_service;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO eye_service;
