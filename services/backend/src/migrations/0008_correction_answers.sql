-- The applicant answers the pending correction requests (ADR-026): they load
-- again what the requests ask for, each value and document with its event,
-- and then send the answer, which answers every request pending at that time.
-- Since 0005, every stored request was pending. From here, a request is
-- pending while no answer covers it: answered_up_to is the number of the last
-- request the answer covered, and the requests after it are still pending.
-- Nothing is deleted, as in the rest of the history.
CREATE TABLE procedure_correction_answers (
    procedure_id   text NOT NULL REFERENCES procedures (id),
    number         integer NOT NULL CHECK (number >= 1),
    answered_up_to integer NOT NULL CHECK (answered_up_to >= 1),
    answered_by    text NOT NULL,
    answered_at    timestamptz NOT NULL,
    PRIMARY KEY (procedure_id, number)
);

-- value_field is the field of a value the applicant entered in answer; the
-- value itself is in procedure_field_values. The answer keeps everything the
-- requests asked for.
ALTER TABLE procedure_events
    ADD COLUMN value_field      text,
    ADD COLUMN answer_fields    text[],
    ADD COLUMN answer_documents text[],
    ADD CONSTRAINT answer_record_is_whole CHECK ((answer_fields IS NULL) = (answer_documents IS NULL));

GRANT SELECT, INSERT ON procedure_correction_answers TO backend;
