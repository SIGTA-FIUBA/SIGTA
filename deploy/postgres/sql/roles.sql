-- Roles and schemas of each service. The passwords are passed with psql -v;
-- see init.sh.

CREATE ROLE backend LOGIN PASSWORD :'backend_password';
CREATE SCHEMA backend AUTHORIZATION backend;
ALTER ROLE backend SET search_path = backend;

CREATE ROLE ai LOGIN PASSWORD :'ai_password';
CREATE SCHEMA ai AUTHORIZATION ai;
ALTER ROLE ai SET search_path = ai;
CREATE EXTENSION IF NOT EXISTS vector SCHEMA ai;
