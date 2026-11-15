-- Roles and schemas of each service. The passwords are passed with psql -v;
-- see init.sh.

-- Up to PostgreSQL 14, like the 11.9 that the faculty offers, every role can
-- create tables in public. Each service creates only in its own schema.
REVOKE CREATE ON SCHEMA public FROM PUBLIC;

-- The backend schema and its tables belong to backend_migrations, which only
-- the migrate command uses. The server connects as backend and gets from each
-- migration only the privileges it needs (ADR-022). NewEmptySchema in
-- services/backend/src/migrations/dbtest copies these owner and privileges
-- for the tests: a change here goes there too.
CREATE ROLE backend_migrations LOGIN PASSWORD :'migrations_password';
CREATE ROLE backend LOGIN PASSWORD :'backend_password';
CREATE SCHEMA backend AUTHORIZATION backend_migrations;
GRANT USAGE ON SCHEMA backend TO backend;
ALTER ROLE backend_migrations SET search_path = backend;
ALTER ROLE backend SET search_path = backend;

CREATE ROLE ai LOGIN PASSWORD :'ai_password';
CREATE SCHEMA ai AUTHORIZATION ai;
ALTER ROLE ai SET search_path = ai;
CREATE EXTENSION IF NOT EXISTS vector SCHEMA ai;
