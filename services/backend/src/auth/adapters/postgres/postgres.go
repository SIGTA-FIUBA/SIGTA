// Package postgres stores the auth module in PostgreSQL, with the tables of
// src/migrations and the privileges of the application role.
package postgres

// Arbitrary numbers, other than those of src/migrations, that identify the
// advisory lock under which each table numbers its new rows.
const (
	areasLockKey       = 7_365_829_106
	membershipsLockKey = 7_365_829_107
)
