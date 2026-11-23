-- The periods in which a person administers areas and memberships (ADR-020).
-- None is deleted: ending one sets valid_until and ended_by, so that the
-- history keeps who granted and who ended each. A person has at most one
-- period without an end.
-- grant_source says how a period was granted. The first administrator of an
-- environment is the initial grant of the system ('bootstrap'), which the
-- grant-admin command adds as the role that owns the tables when there is no
-- administrator, and the only kind without granted_by: SIGTA stores no author
-- that someone typed, and keeping who ran the command is up to the environment
-- where it is deployed. The others are
-- granted by an administrator ('administrator'), the user of the session that
-- granted them.
-- The server only reads them, so that a session cannot make itself an
-- administrator. The migration that lets administrators grant and end them in
-- SIGTA gives the application role what that needs.
-- There is no reference to users, as in the rest of the auth module.
CREATE TABLE administrators (
    id           text PRIMARY KEY,
    user_id      text NOT NULL,
    valid_from   timestamptz NOT NULL,
    valid_until  timestamptz,
    grant_source text NOT NULL CHECK (grant_source IN ('bootstrap', 'administrator')),
    granted_by   text,
    ended_by     text,
    CHECK ((grant_source = 'bootstrap') = (granted_by IS NULL)),
    CHECK ((valid_until IS NULL) = (ended_by IS NULL)),
    CHECK (valid_until IS NULL OR valid_until > valid_from)
);

CREATE UNIQUE INDEX administrators_without_end ON administrators (user_id) WHERE valid_until IS NULL;

GRANT SELECT ON administrators TO backend;
