package rest

import (
	"context"

	"backend/src/api"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/ports/input"
)

type CatalogHandler struct {
	procedureTypes input.ProcedureTypes
}

func NewCatalogHandler(procedureTypes input.ProcedureTypes) *CatalogHandler {
	return &CatalogHandler{procedureTypes: procedureTypes}
}

func (h *CatalogHandler) ListProcedureTypes(ctx context.Context, _ api.ListProcedureTypesRequestObject) (api.ListProcedureTypesResponseObject, error) {
	entries, err := h.procedureTypes.Catalog(ctx)
	if err != nil {
		return nil, err
	}
	response := make(api.ListProcedureTypes200JSONResponse, 0, len(entries))
	for _, entry := range entries {
		response = append(response, catalogEntry(entry))
	}
	return response, nil
}

func catalogEntry(entry input.CatalogEntry) api.CatalogEntry {
	return api.CatalogEntry{
		Id:            entry.ProcedureType,
		Version:       int64(entry.Version),
		Name:          entry.Name,
		Description:   entry.Description,
		Deadlines:     orEmpty(entry.Deadlines),
		Admissibility: orEmpty(entry.Admissibility),
		Fields:        fieldRequirements(entry.Fields),
		Documents:     documentRequirements(entry.Documents),
		Submission: api.SubmissionForm{
			Fields:    fieldRequirements(entry.SubmissionFields),
			Documents: documentRequirements(entry.SubmissionDocuments),
		},
	}
}

func fieldRequirements(requirements []definition.FieldRequirement) []api.FieldRequirement {
	fields := make([]api.FieldRequirement, 0, len(requirements))
	for _, r := range requirements {
		fields = append(fields, fieldRequirement(r))
	}
	return fields
}

func documentRequirements(requirements []definition.DocumentRequirement) []api.DocumentRequirement {
	documents := make([]api.DocumentRequirement, 0, len(requirements))
	for _, r := range requirements {
		documents = append(documents, documentRequirement(r))
	}
	return documents
}

func fieldRequirement(r definition.FieldRequirement) api.FieldRequirement {
	options := make([]api.NamedOption, 0, len(r.Field.Options))
	for _, o := range r.Field.Options {
		options = append(options, api.NamedOption{Id: o.ID, Name: o.Name})
	}
	return api.FieldRequirement{Id: r.Field.ID, Name: r.Field.Name, Required: r.Required, Type: api.FieldType(r.Field.Type), Options: options}
}

func documentRequirement(r definition.DocumentRequirement) api.DocumentRequirement {
	return api.DocumentRequirement{
		Id:            r.Document.ID,
		Name:          r.Document.Name,
		Required:      r.Required,
		AcceptedTypes: orEmpty(r.Document.AcceptedTypes),
		MaxSize:       r.Document.MaxSizeBytes,
		Resolution:    r.Document.Resolution,
	}
}

// orEmpty keeps required lists as [] in the JSON, never null.
func orEmpty(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
