-- Areas, their permissions and roles, and memberships (ADR-020). None of
-- them is deleted: deleting an area, a permission or a role marks it with
-- deleted_at, and ending a membership sets valid_until, so that the history
-- still names them. The application role cannot delete their rows.
-- number keeps the order in which they were added, which is the order the
-- application lists them in. There are no sequences, so it is computed when
-- the row is inserted.
-- deleted_at can go back to NULL: SIGTA restores a platform permission of the
-- Secretariat that was deleted by hand.
CREATE TABLE areas (
    id         text PRIMARY KEY,
    number     integer NOT NULL UNIQUE CHECK (number >= 1),
    name       text NOT NULL,
    system     boolean NOT NULL,
    deleted_at timestamptz,
    deleted_by text
);

CREATE TABLE area_permissions (
    area_id    text NOT NULL REFERENCES areas (id),
    id         text NOT NULL,
    number     integer NOT NULL CHECK (number >= 1),
    name       text NOT NULL,
    system     boolean NOT NULL,
    deleted_at timestamptz,
    deleted_by text,
    PRIMARY KEY (area_id, id),
    UNIQUE (area_id, number)
);

CREATE TABLE area_roles (
    area_id    text NOT NULL REFERENCES areas (id),
    id         text NOT NULL,
    number     integer NOT NULL CHECK (number >= 1),
    name       text NOT NULL,
    deleted_at timestamptz,
    deleted_by text,
    PRIMARY KEY (area_id, id),
    UNIQUE (area_id, number)
);

-- A change of the permissions of a role is a new revision, and the earlier
-- ones tell what the role gave at each time, so they never change.
CREATE TABLE area_role_revisions (
    area_id     text NOT NULL,
    role_id     text NOT NULL,
    number      integer NOT NULL CHECK (number >= 1),
    permissions text[] NOT NULL,
    revised_at  timestamptz NOT NULL,
    revised_by  text NOT NULL,
    PRIMARY KEY (area_id, role_id, number),
    FOREIGN KEY (area_id, role_id) REFERENCES area_roles (area_id, id)
);

-- There are no references to users or areas, as in the rest of the auth
-- module: rows are not deleted, so the application checks that what they
-- name exists. Ending a membership that has not started yet also moves its
-- start, so that it ends after it starts.
CREATE TABLE memberships (
    id          text PRIMARY KEY,
    number      integer NOT NULL UNIQUE CHECK (number >= 1),
    user_id     text NOT NULL,
    area_id     text NOT NULL,
    roles       text[] NOT NULL,
    permissions text[] NOT NULL,
    scope       bigint[] NOT NULL,
    valid_from  timestamptz NOT NULL,
    valid_until timestamptz,
    granted_by  text NOT NULL,
    ended_by    text
);

GRANT SELECT, INSERT ON areas, area_permissions, area_roles, area_role_revisions, memberships TO backend;
GRANT UPDATE (name, system, deleted_at, deleted_by) ON areas, area_permissions TO backend;
GRANT UPDATE (name, deleted_at, deleted_by) ON area_roles TO backend;
GRANT UPDATE (valid_from, valid_until, ended_by) ON memberships TO backend;
