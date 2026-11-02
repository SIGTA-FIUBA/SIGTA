package rest

import (
	"context"

	"backend/src/api"
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
		documents = append(documents, api.DocumentRequirement{
			Id:            document.Document.ID,
			Name:          document.Document.Name,
			Required:      document.Required,
			AcceptedTypes: orEmpty(document.Document.AcceptedTypes),
			MaxSize:       document.Document.MaxSize,
		})
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

// orEmpty keeps required lists as [] in the JSON, never null.
func orEmpty(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
