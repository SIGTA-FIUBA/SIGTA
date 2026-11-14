-- The draft of a procedure type that is edited before it is published
-- (ADR-013). Unlike a version, a draft changes and is deleted once it is
-- published or discarded, so the application role can update and delete it.
-- It cannot move a draft to another procedure type or base: a draft from
-- another version is a new draft.
-- definition is jsonb: nothing covers its bytes with a digest, as it does
-- with a version.
-- There is no reference to procedure_types, since that row is created with
-- version 1 and the draft of a new procedure type comes before it.
CREATE TABLE procedure_type_drafts (
    procedure_type_id bigint PRIMARY KEY CHECK (procedure_type_id >= 1),
    base              integer NOT NULL CHECK (base >= 0),
    definition        jsonb NOT NULL,
    status            text NOT NULL CHECK (status IN ('editing', 'proposed')),
    revision          integer NOT NULL CHECK (revision >= 1),
    edited_by         text NOT NULL,
    edited_at         timestamptz NOT NULL,
    proposed_by       text,
    proposed_at       timestamptz,
    rejected_by       text,
    rejected_at       timestamptz,
    rejection_reason  text,
    CONSTRAINT proposed_has_proposer CHECK ((status = 'proposed') = (proposed_by IS NOT NULL AND proposed_at IS NOT NULL))
);

GRANT SELECT, INSERT, DELETE ON procedure_type_drafts TO backend;
GRANT UPDATE (definition, status, revision, edited_by, edited_at, proposed_by, proposed_at, rejected_by, rejected_at, rejection_reason)
    ON procedure_type_drafts TO backend;
