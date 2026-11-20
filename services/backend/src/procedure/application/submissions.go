package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/domain/submission"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

var _ input.Submissions = (*SubmissionService)(nil)

// SubmissionService starts procedures with what their opening steps ask for.
// Anyone with an account can start one in their own name, staff included.
type SubmissionService struct {
	procedures *ProcedureService
	files      output.Files
	newID      func() string
	now        func() time.Time
}

func NewSubmissionService(procedures *ProcedureService, files output.Files, newID func() string, now func() time.Time) *SubmissionService {
	return &SubmissionService{procedures: procedures, files: files, newID: newID, now: now}
}

func (s *SubmissionService) Submit(ctx context.Context, userID string, sub input.Submission) (input.Started, error) {
	latest, err := s.procedures.latest(ctx, sub.ProcedureType)
	if err != nil {
		return input.Started{}, err
	}
	content, err := latest.DecodeContent()
	if err != nil {
		return input.Started{}, fmt.Errorf("procedure type %d version %d: %w", sub.ProcedureType, latest.Number, err)
	}
	def := content.Definition
	// The values come before the files, so a submission with a wrong value is
	// rejected without storing any file.
	values, problems := submission.CheckValues(def, sub.Values)
	if err := submission.Reject(problems); err != nil {
		return input.Started{}, err
	}

	p := domain.Procedure{
		ID:            s.newID(),
		ProcedureType: sub.ProcedureType,
		Version:       latest.Number,
		VersionDigest: latest.Digest,
		Applicant:     userID,
		OpenStep:      submission.OpenStep(def),
	}
	now := s.now()
	p.StartedAt, p.MovedAt = now, now
	state, err := domain.StateOf(p, def)
	if err != nil {
		return input.Started{}, err
	}
	fields, requirements := submission.Requirements(def)
	documents, err := s.storeFiles(ctx, p.ID, userID, now, requirements, sub)
	if err != nil {
		return input.Started{}, err
	}

	var entered []domain.FieldValue
	for _, r := range fields {
		v, ok := values[r.Field.ID]
		if !ok {
			continue
		}
		entered = append(entered, domain.FieldValue{
			Procedure: p.ID,
			Field:     r.Field.ID,
			Step:      r.Step,
			Text:      v.Text,
			Items:     v.Items,
			EnteredBy: userID,
			EnteredAt: now,
		})
	}
	// A failure here leaves files that no document names, which
	// RemoveUnreferencedFiles removes at startup.
	start := output.Start{
		Procedure: p,
		Events:    engine.Open(p, submission.OpeningSteps(def), userID, now),
		Values:    entered,
		Documents: documents,
	}
	if err := s.procedures.procedures.Add(ctx, start); err != nil {
		return input.Started{}, fmt.Errorf("start procedure: %w", err)
	}
	return input.Started{Procedure: p, State: state}, nil
}

// storeFiles keeps every file that covers a document the opening steps ask
// for, and returns the documents that name them, still to be recorded.
func (s *SubmissionService) storeFiles(ctx context.Context, procedureID, userID string, now time.Time, requirements []submission.StepDocumentRequirement, sub input.Submission) ([]domain.Document, error) {
	asked := make(map[string]submission.StepDocumentRequirement, len(requirements))
	for _, r := range requirements {
		asked[r.Document.ID] = r
	}
	var (
		problems  []submission.Problem
		documents []domain.Document
		sent      = map[string]bool{}
	)
	if sub.Files != nil {
		for f, err := range sub.Files {
			if err != nil {
				return nil, err
			}
			r, ok := asked[f.Requirement]
			switch {
			case !ok:
				problems = append(problems, submission.DocumentProblem(submission.CodeNotAsked, f.Requirement))
				continue
			case sent[f.Requirement]:
				problems = append(problems, submission.DocumentProblem(submission.CodeRepeated, f.Requirement))
				continue
			}
			sent[f.Requirement] = true
			stored, err := storeFile(ctx, s.files, r.Document, f.FileName, f.Content)
			if code, rejected := fileProblem(err); rejected {
				problems = append(problems, submission.DocumentProblem(code, f.Requirement))
				continue
			}
			if err != nil {
				return nil, err
			}
			documents = append(documents, domain.Document{
				ID:          s.newID(),
				Procedure:   procedureID,
				Requirement: r.Document.ID,
				Step:        r.Step,
				FileName:    stored.fileName,
				MediaType:   stored.mediaType,
				Size:        stored.size,
				Digest:      stored.digest,
				UploadedBy:  userID,
				UploadedAt:  now,
			})
		}
	}
	problems = append(problems, submission.MissingDocuments(requirements, sent)...)
	if err := submission.Reject(problems); err != nil {
		return nil, err
	}
	return documents, nil
}

func fileProblem(err error) (submission.ProblemCode, bool) {
	switch {
	case errors.Is(err, domain.ErrDocumentEmpty):
		return submission.CodeEmpty, true
	case errors.Is(err, domain.ErrDocumentTypeNotAccepted):
		return submission.CodeTypeNotAccepted, true
	case errors.Is(err, domain.ErrDocumentTooLarge):
		return submission.CodeTooLarge, true
	}
	return "", false
}
