package rest

import (
	"context"

	ports "backend/src/ports/input"
)

type CatalogHandler struct {
	procedureTypes ports.ProcedureTypes
}

func NewCatalogHandler(procedureTypes ports.ProcedureTypes) *CatalogHandler {
	return &CatalogHandler{procedureTypes: procedureTypes}
}

func (h *CatalogHandler) ListProcedureTypes(ctx context.Context, _ ListProcedureTypesRequestObject) (ListProcedureTypesResponseObject, error) {
	entries, err := h.procedureTypes.Catalog(ctx)
	if err != nil {
		return nil, err
	}
	response := make(ListProcedureTypes200JSONResponse, 0, len(entries))
	for _, entry := range entries {
		response = append(response, catalogEntry(entry))
	}
	return response, nil
}

func catalogEntry(entry ports.CatalogEntry) CatalogEntry {
	fields := make([]FieldRequirement, 0, len(entry.Fields))
	for _, f := range entry.Fields {
		fields = append(fields, FieldRequirement{Id: f.Field.ID, Name: f.Field.Name, Required: f.Required})
	}
	documents := make([]DocumentRequirement, 0, len(entry.Documents))
	for _, d := range entry.Documents {
		documents = append(documents, DocumentRequirement{
			Id:            d.Document.ID,
			Name:          d.Document.Name,
			Required:      d.Required,
			AcceptedTypes: orEmpty(d.Document.AcceptedTypes),
			MaxSize:       d.Document.MaxSize,
		})
	}
	return CatalogEntry{
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
