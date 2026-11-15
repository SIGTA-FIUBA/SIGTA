package rest

import (
	"context"

	"backend/src/api"
	"backend/src/procedure/domain/definition"
	ports "backend/src/procedure/ports/input"
)

type CatalogHandler struct {
	procedureTypes ports.ProcedureTypes
}

func NewCatalogHandler(procedureTypes ports.ProcedureTypes) *CatalogHandler {
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

func catalogEntry(entry ports.CatalogEntry) api.CatalogEntry {
	fields := make([]api.FieldRequirement, 0, len(entry.Fields))
	for _, field := range entry.Fields {
		fields = append(fields, api.FieldRequirement{Id: field.Field.ID, Name: field.Field.Name, Required: field.Required})
	}
	documents := make([]api.DocumentRequirement, 0, len(entry.Documents))
	for _, document := range entry.Documents {
		documents = append(documents, documentRequirement(document))
	}
	return api.CatalogEntry{
		Id:            entry.ProcedureType,
		Version:       entry.Version,
		Name:          entry.Name,
		Description:   entry.Description,
		Deadlines:     orEmpty(entry.Deadlines),
		Admissibility: orEmpty(entry.Admissibility),
		Fields:        fields,
		Documents:     documents,
	}
}

func documentRequirement(r definition.DocumentRequirement) api.DocumentRequirement {
	return api.DocumentRequirement{
		Id:            r.Document.ID,
		Name:          r.Document.Name,
		Required:      r.Required,
		AcceptedTypes: orEmpty(r.Document.AcceptedTypes),
		MaxSize:       r.Document.MaxSizeBytes,
	}
}

// orEmpty keeps required lists as [] in the JSON, never null.
func orEmpty(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
