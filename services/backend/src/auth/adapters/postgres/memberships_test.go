package postgres_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"backend/src/auth/adapters/postgres"
	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
	"backend/src/auth/ports/output/outputtest"
	"backend/src/migrations/dbtest"
)

// office is an area that the memberships of the contract are granted in,
// which the repository asks to exist, as committee.
func office(at time.Time) domain.Area {
	return domain.Area{ID: "office", Name: "Mesa", Permissions: []domain.Permission{
		{ID: "approve", Name: "Aprobar"}, {ID: "review", Name: "Revisar"},
	}, Roles: []domain.Role{
		{ID: "reviewer", Name: "Revisión", Revisions: []domain.RoleRevision{{Permissions: []string{"review"}, At: at, By: "admin"}}},
	}}
}

func TestMembershipRepository(t *testing.T) {
	for _, isolation := range []dbtest.Isolation{dbtest.ReadCommitted, dbtest.Serializable} {
		t.Run(string(isolation), func(t *testing.T) {
			outputtest.TestMembershipRepository(t, func(t *testing.T) output.MembershipRepository {
				pool := dbtest.NewSchema(t).Pool(t, dbtest.AppRole, isolation)
				committee := office(time.Now())
				committee.ID, committee.Name = "committee", "Comisión"
				for _, a := range []domain.Area{office(time.Now()), committee} {
					if err := postgres.NewAreaRepository(pool).Save(t.Context(), a); err != nil {
						t.Fatal(err)
					}
				}
				return postgres.NewMembershipRepository(pool)
			})
		})
	}
}

// Removing an area, a permission or a role while it is granted leaves either
// the grant or the removal, never a membership in what was removed.
func TestARemovalAndAGrantAtOnceDoNotBothPass(t *testing.T) {
	at := time.Date(2030, 1, 1, 12, 0, 0, 0, time.UTC)
	for name, remove := range map[string]func(*domain.Area){
		"area":       func(a *domain.Area) { a.DeletedAt, a.DeletedBy = at, "admin" },
		"permission": func(a *domain.Area) { a.Permissions[1].DeletedAt, a.Permissions[1].DeletedBy = at, "admin" },
		"role":       func(a *domain.Area) { a.Roles[0].DeletedAt, a.Roles[0].DeletedBy = at, "admin" },
	} {
		t.Run(name, func(t *testing.T) {
			pool := dbtest.NewSchema(t).Pool(t, dbtest.AppRole, dbtest.ReadCommitted)
			areas, memberships := postgres.NewAreaRepository(pool), postgres.NewMembershipRepository(pool)
			area := office(at)
			if err := areas.Save(t.Context(), area); err != nil {
				t.Fatal(err)
			}
			removed := office(at)
			remove(&removed)
			for range 20 {
				var grant, removal error
				var wg sync.WaitGroup
				wg.Go(func() {
					grant = memberships.Save(t.Context(), domain.Membership{
						ID: "M-1", UserID: "U-1", Area: "office", Roles: []string{"reviewer"}, Permissions: []string{"review"},
						From: at.Add(-time.Hour), GrantedBy: "admin",
					})
				})
				wg.Go(func() { removal = areas.Save(t.Context(), removed) })
				wg.Wait()
				if (grant == nil) == (removal == nil) {
					t.Fatalf("grant = %v, removal = %v; want exactly one to fail", grant, removal)
				}
				if grant == nil {
					return // The grant went first and the removal saw it.
				}
				if !errors.Is(grant, domain.ErrInvalidMembership) {
					t.Fatalf("grant = %v, want ErrInvalidMembership", grant)
				}
				// The removal went first: start again until the grant does.
				pool = dbtest.NewSchema(t).Pool(t, dbtest.AppRole, dbtest.ReadCommitted)
				areas, memberships = postgres.NewAreaRepository(pool), postgres.NewMembershipRepository(pool)
				if err := areas.Save(t.Context(), area); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
