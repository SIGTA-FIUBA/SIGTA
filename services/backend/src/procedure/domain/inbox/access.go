package inbox

import "backend/src/procedure/domain/definition"

// Person is who asks, as the auth module knows them.
type Person struct {
	ID   string
	Role string
}

// Access is what a person can see of the procedures in progress. An applicant
// sees only their own procedures. A staff member sees the procedure types they
// hold a profile in, and that profile decides which steps they can do.
//
// Who holds which profile in which type comes from the permissions, which the
// inbox reads through a port and does not decide.
type Access struct {
	// Applicant is set for an applicant: they see only the procedures they
	// started, whatever the other fields say.
	Applicant string
	// Profiles is the profile a staff member holds in each procedure type.
	Profiles map[int64]definition.Profile
	// EveryType is a profile held in every procedure type. It stands in for
	// per-type permissions until they exist, and is empty once they do.
	EveryType definition.Profile
}

// ProfileIn is the profile held in a procedure type, or "" if none: then the
// type is not visible.
func (a Access) ProfileIn(procedureType int64) definition.Profile {
	if a.Applicant != "" {
		return ""
	}
	if profile := a.Profiles[procedureType]; profile != "" {
		return profile
	}
	return a.EveryType
}

// CanSee tells whether a procedure is visible with this access.
func (a Access) CanSee(s Summary) bool {
	if a.Applicant != "" {
		return s.ApplicantID == a.Applicant
	}
	return a.ProfileIn(s.ProcedureType) != ""
}

// Staff tells whether the access is a staff member's, who may see procedures
// of others.
func (a Access) Staff() bool {
	return a.Applicant == "" && (len(a.Profiles) > 0 || a.EveryType != "")
}
