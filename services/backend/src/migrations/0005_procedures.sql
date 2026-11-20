-- What a procedure records as it moves: where it is, the history of what was
-- done, the values entered and the documents attached. Starting a procedure
-- writes the procedure, its values, its documents and its first events in one
-- transaction, and attaching a document writes the document and its event in
-- another.
-- No procedure was stored before this migration, so the new columns are NOT
-- NULL without a default.
-- An empty value of the domain is NULL: open_step when the procedure is
-- closed, next_step when the event does not move the procedure.
ALTER TABLE procedures
    ADD COLUMN applicant      text NOT NULL,
    ADD COLUMN open_step      text,
    ADD COLUMN referral_from  text,
    ADD COLUMN referral_to    text,
    ADD COLUMN referral_since timestamptz,
    ADD COLUMN started_at     timestamptz NOT NULL,
    ADD COLUMN moved_at       timestamptz NOT NULL,
    ADD CONSTRAINT referral_is_whole CHECK (
        (referral_from IS NULL) = (referral_to IS NULL) AND (referral_from IS NULL) = (referral_since IS NULL));

-- number is the order of the documents within their procedure, computed when
-- the row is inserted, since there are no sequences.
CREATE TABLE documents (
    id           text PRIMARY KEY,
    procedure_id text NOT NULL REFERENCES procedures (id),
    number       integer NOT NULL CHECK (number >= 1),
    requirement  text NOT NULL,
    step         text NOT NULL,
    file_name    text NOT NULL,
    media_type   text NOT NULL,
    size         bigint NOT NULL CHECK (size >= 0),
    digest       text NOT NULL,
    uploaded_by  text NOT NULL,
    uploaded_at  timestamptz NOT NULL,
    UNIQUE (procedure_id, number),
    -- Lets an event name a document of its own procedure only.
    UNIQUE (procedure_id, id)
);

-- The event keeps the name of the document as it was when it was attached,
-- as the domain does, along with the reference.
ALTER TABLE procedure_events
    ADD COLUMN step                 text NOT NULL,
    ADD COLUMN action               text NOT NULL,
    ADD COLUMN done_by              text NOT NULL,
    ADD COLUMN done_at              timestamptz NOT NULL,
    ADD COLUMN observation          text,
    ADD COLUMN outcome              text,
    ADD COLUMN declared             boolean NOT NULL,
    ADD COLUMN send_back            boolean NOT NULL,
    ADD COLUMN next_step            text,
    ADD COLUMN referral_from        text,
    ADD COLUMN referral_to          text,
    ADD COLUMN external_file_number text,
    ADD COLUMN returned             boolean NOT NULL,
    ADD COLUMN document_id          text,
    ADD COLUMN document_requirement text,
    ADD COLUMN document_file_name   text,
    ADD CONSTRAINT referral_record_is_whole CHECK (
        (referral_from IS NULL) = (referral_to IS NULL) AND (external_file_number IS NULL OR referral_from IS NOT NULL)),
    ADD CONSTRAINT document_record_is_whole CHECK (
        (document_id IS NULL) = (document_requirement IS NULL) AND (document_id IS NULL) = (document_file_name IS NULL)),
    ADD FOREIGN KEY (procedure_id, document_id) REFERENCES documents (procedure_id, id) DEFERRABLE INITIALLY DEFERRED;

CREATE TABLE procedure_field_values (
    procedure_id text NOT NULL REFERENCES procedures (id),
    number       integer NOT NULL CHECK (number >= 1),
    field        text NOT NULL,
    step         text NOT NULL,
    text         text NOT NULL,
    items        text[] NOT NULL,
    entered_by   text NOT NULL,
    entered_at   timestamptz NOT NULL,
    PRIMARY KEY (procedure_id, number)
);

GRANT SELECT, INSERT ON documents, procedure_field_values TO backend;
-- Moving a procedure changes only where it is. Locking its row, which the
-- adapter does to number what it appends, also needs UPDATE on some column.
GRANT UPDATE (open_step, referral_from, referral_to, referral_since, moved_at) ON procedures TO backend;
