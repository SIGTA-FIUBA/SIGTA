package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"backend/src/api"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain/submission"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

const (
	messageSubmissionRejected = "El trámite no cumple con lo que pide. Revisá los datos y documentos marcados."
	messageNoProcedureType    = "No existe el tipo de trámite."

	formProcedureType = "procedure_type"
	formValues        = "values"
	// maxValuesBytes bounds the JSON of the values, far above what the
	// largest fields allow.
	maxValuesBytes = 1 << 20
	maxTypeLength  = 32
)

// messageSubmissionTooLarge names api.MaxRequestBytes, the most one request
// carries, whatever each document allows.
var messageSubmissionTooLarge = fmt.Sprintf("Los archivos suman más de lo que se puede enviar de una vez (%d MB). Achicá alguno y volvé a intentarlo.", api.MaxRequestBytes>>20)

func (h *ProcedureHandler) SubmitProcedure(ctx context.Context, req api.SubmitProcedureRequestObject) (api.SubmitProcedureResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.SubmitProcedure401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	procedureType, values, err := submissionHead(req.Body)
	if err != nil {
		return malformedSubmission(), nil
	}
	started, err := h.submissions.Submit(ctx, user.ID, input.Submission{
		ProcedureType: procedureType,
		Values:        values,
		Files:         submittedFiles(req.Body),
	})
	var rejected *submission.Error
	var requestTooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &rejected):
		return api.SubmitProcedure400JSONResponse{Message: messageSubmissionRejected, Problems: submissionProblems(rejected.Problems)}, nil
	// Before the malformed upload, which wraps it when the request is cut
	// while reading a file.
	case errors.As(err, &requestTooLarge):
		return api.SubmitProcedure400JSONResponse{Message: messageSubmissionTooLarge, Problems: []api.SubmissionProblem{}}, nil
	case errors.Is(err, errMalformedUpload):
		return malformedSubmission(), nil
	case errors.Is(err, output.ErrVersionNotFound):
		return api.SubmitProcedure404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageNoProcedureType}}, nil
	case err != nil:
		return nil, err
	}
	p := started.Procedure
	response := api.SubmitProcedure201JSONResponse{Id: p.ID, State: api.ProcedureState(started.State)}
	if !p.Closed() {
		response.OpenStep = &p.OpenStep
	}
	return response, nil
}

func malformedSubmission() api.SubmitProcedure400JSONResponse {
	return api.SubmitProcedure400JSONResponse{Message: api.InvalidRequestMessage, Problems: []api.SubmissionProblem{}}
}

// submissionHead reads the procedure type and the values, the first two
// parts, and leaves the files unread.
func submissionHead(body *multipart.Reader) (int64, map[string]submission.Value, error) {
	typePart, err := formPart(body, formProcedureType, maxTypeLength)
	if err != nil {
		return 0, nil, err
	}
	procedureType, err := strconv.ParseInt(string(typePart), 10, 64)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: %w", errMalformedUpload, err)
	}
	valuesPart, err := formPart(body, formValues, maxValuesBytes)
	if err != nil {
		return 0, nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(valuesPart, &raw); err != nil {
		return 0, nil, fmt.Errorf("%w: %w", errMalformedUpload, err)
	}
	values := make(map[string]submission.Value, len(raw))
	for id, r := range raw {
		var text string
		if err := json.Unmarshal(r, &text); err == nil {
			values[id] = submission.Value{Text: text}
			continue
		}
		var items []string
		if err := json.Unmarshal(r, &items); err != nil || items == nil {
			return 0, nil, fmt.Errorf("%w: value of %s is neither a text nor a list of texts", errMalformedUpload, id)
		}
		values[id] = submission.Value{Items: items}
	}
	// PostgreSQL cannot store a text with a NUL character.
	for id, v := range values {
		if strings.ContainsRune(id+v.Text+strings.Join(v.Items, ""), 0) {
			return 0, nil, fmt.Errorf("%w: value of %q has a NUL character", errMalformedUpload, id)
		}
	}
	return procedureType, values, nil
}

func formPart(body *multipart.Reader, name string, maxBytes int64) ([]byte, error) {
	part, err := body.NextPart()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errMalformedUpload, err)
	}
	if part.FormName() != name {
		return nil, fmt.Errorf("%w: %s instead of %s", errMalformedUpload, part.FormName(), name)
	}
	value, err := io.ReadAll(io.LimitReader(part, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errMalformedUpload, err)
	}
	if int64(len(value)) > maxBytes {
		return nil, fmt.Errorf("%w: %s is too long", errMalformedUpload, name)
	}
	return value, nil
}

// submittedFiles yields each pair of document and file parts until the end
// of the request. The use case reads each file before asking for the next.
func submittedFiles(body *multipart.Reader) iter.Seq2[input.SubmittedFile, error] {
	return func(yield func(input.SubmittedFile, error) bool) {
		for {
			requirement, file, err := uploadParts(body)
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				yield(input.SubmittedFile{}, err)
				return
			}
			if !yield(input.SubmittedFile{Requirement: requirement, FileName: file.FileName(), Content: partReader{file}}, nil) {
				return
			}
		}
	}
}

func submissionProblems(problems []submission.Problem) []api.SubmissionProblem {
	response := make([]api.SubmissionProblem, 0, len(problems))
	for _, p := range problems {
		response = append(response, api.SubmissionProblem{
			Code: api.SubmissionProblemCode(p.Code),
			Kind: api.SubmissionProblemKind(p.Kind),
			Id:   p.ID,
		})
	}
	return response
}
