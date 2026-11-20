-- Tables of the auth module for people and their sessions. A user is never
-- deleted nor changed: the application role only reads and adds them.
-- The email is unique, so two first logins of the same person at once
-- register only one user.
CREATE TABLE users (
    id    text PRIMARY KEY,
    email text NOT NULL UNIQUE,
    name  text NOT NULL
);

-- A session is not history: it is deleted when it is closed or used after
-- it expired, and only the time it was last used changes. One that is never
-- used again stays. There is no reference to users,
-- as in the rest of the auth module: rows are not deleted, so the
-- application checks that what they name exists.
CREATE TABLE sessions (
    token_hash   text PRIMARY KEY,
    user_id      text NOT NULL,
    last_seen_at timestamptz NOT NULL
);

GRANT SELECT, INSERT ON users TO backend;
GRANT SELECT, INSERT, DELETE ON sessions TO backend;
GRANT UPDATE (last_seen_at) ON sessions TO backend;
