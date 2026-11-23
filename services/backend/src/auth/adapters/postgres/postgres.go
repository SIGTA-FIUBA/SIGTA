// Package postgres stores the auth module in PostgreSQL, with the tables of
// src/migrations and the privileges of the application role.
package postgres

// Arbitrary numbers, other than those of src/migrations, that identify the
// advisory lock under which each table numbers its new rows, or, for the
// administrators, finds whether there is one.
const (
	areasLockKey          = 7_365_829_106
	membershipsLockKey    = 7_365_829_107
	administratorsLockKey = 7_365_829_108
)

const uniqueViolation = "23505"
