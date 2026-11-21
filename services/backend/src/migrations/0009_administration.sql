-- What administrators need to manage areas, memberships and the other
-- administrators in SIGTA (US74), with who did each change and when.
-- An area or a permission records who created it; those that come with SIGTA,
-- and those created before this migration, have no author. A membership
-- records when it was granted, which valid_from may follow.
ALTER TABLE areas ADD COLUMN created_at timestamptz, ADD COLUMN created_by text,
    ADD CHECK ((created_at IS NULL) = (created_by IS NULL));
ALTER TABLE area_permissions ADD COLUMN created_at timestamptz, ADD COLUMN created_by text,
    ADD CHECK ((created_at IS NULL) = (created_by IS NULL));
ALTER TABLE memberships ADD COLUMN granted_at timestamptz;

-- The application role grants and ends administrators, but never as the
-- initial grant of the system, which stays with grant-admin, and it only ends
-- an open period, never reopens or moves one. Row level security does not
-- bind the role that owns the table, which grant-admin connects as.
ALTER TABLE administrators ENABLE ROW LEVEL SECURITY;
CREATE POLICY read ON administrators FOR SELECT TO backend USING (true);
CREATE POLICY grant_by_administrator ON administrators FOR INSERT TO backend
    WITH CHECK (grant_source = 'administrator');
CREATE POLICY end_open_period ON administrators FOR UPDATE TO backend
    USING (valid_until IS NULL) WITH CHECK (valid_until IS NOT NULL);
GRANT INSERT ON administrators TO backend;
GRANT UPDATE (valid_until, ended_by) ON administrators TO backend;
