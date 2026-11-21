package migrations

const LockKey = lockKey

var (
	Files     = files
	Load      = load
	ApplyFrom = apply
)

// AppendOnlyTables returns the immutable tables that have no column to update.
func AppendOnlyTables() []string {
	var tables []string
	for _, t := range immutable {
		if len(t.columns) == 0 {
			tables = append(tables, t.table)
		}
	}
	return tables
}
