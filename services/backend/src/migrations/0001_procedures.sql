-- Tables of the procedure module. They belong to the role that runs the
-- migrations: the owner of a table can always change or drop it, so the
-- application role only gets the privileges it needs, and none to change a
-- published version, the version a procedure follows or an event. Nothing
-- updates these tables yet, so the application role cannot update any of
-- them.
-- Only what PostgreSQL 11 has is used, since the faculty runs that version.

GRANT SELECT ON schema_migrations TO backend;

-- The row is created along with version 1 of the procedure type.
CREATE TABLE procedure_types (
    id bigint PRIMARY KEY
);

-- content is bytea and not jsonb: jsonb reorders keys and would change the
-- bytes the digest covers. There is no format column, since the format is
-- inside the content.
CREATE TABLE procedure_type_versions (
    procedure_type_id bigint NOT NULL REFERENCES procedure_types (id),
    number            integer NOT NULL CHECK (number >= 1),
    digest            text NOT NULL,
    content           bytea NOT NULL,
    proposed_by       text NOT NULL,
    approved_by       text NOT NULL,
    published_at      timestamptz NOT NULL,
    PRIMARY KEY (procedure_type_id, number),
    -- Lets procedures reference the digest they pinned along with the number.
    UNIQUE (procedure_type_id, number, digest),
    -- lower() because Authorship.Validate compares without case. lower()
    -- depends on the locale of the database and strings.EqualFold does not,
    -- so they can disagree outside ASCII: the adapter reports this
    -- constraint as an invalid authorship.
    CONSTRAINT approver_is_not_proposer CHECK (lower(approved_by) <> lower(proposed_by))
);

CREATE VIEW current_procedure_type_versions AS
SELECT DISTINCT ON (procedure_type_id) *
FROM procedure_type_versions
ORDER BY procedure_type_id, number DESC;

-- A procedure keeps the digest of its version as well as the number, so that
-- a lost version whose number is published again is noticed. The rest of its
-- columns come with the stories that start and advance a procedure.
CREATE TABLE procedures (
    id                text PRIMARY KEY,
    procedure_type_id bigint NOT NULL,
    version_number    integer NOT NULL,
    version_digest    text NOT NULL,
    FOREIGN KEY (procedure_type_id, version_number, version_digest)
        REFERENCES procedure_type_versions (procedure_type_id, number, digest)
);

-- An event is never changed: what happens later is a new event.
CREATE TABLE procedure_events (
    procedure_id text NOT NULL REFERENCES procedures (id),
    number       integer NOT NULL CHECK (number >= 1),
    PRIMARY KEY (procedure_id, number)
);

GRANT SELECT, INSERT ON procedure_types, procedure_type_versions, procedures, procedure_events TO backend;
GRANT SELECT ON current_procedure_type_versions TO backend;
