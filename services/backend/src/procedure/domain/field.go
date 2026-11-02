package domain

import (
	"mime"
	"strings"
)

type FieldType string

const (
	FieldText        FieldType = "text"
	FieldLongText    FieldType = "long_text"
	FieldNumber      FieldType = "number"
	FieldDate        FieldType = "date"
	FieldOption      FieldType = "option"
	FieldDegree      FieldType = "degree"
	FieldSubjectList FieldType = "subject_list"
)

type Field struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Type    FieldType `json:"type"`
	Options []Option  `json:"options,omitempty"`
}

// Option is stored by its id, so that correcting its name later does not
// change what earlier answers mean.
type Option struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Document struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// AcceptedTypes are media types, such as "application/pdf".
	AcceptedTypes []string `json:"accepted_types,omitempty"`
	// MaxSize is in bytes.
	MaxSize int64 `json:"max_size,omitempty"`
}

func validateFields(fields []Field, seen map[string]bool) []Problem {
	var problems []Problem
	for _, f := range fields {
		problems = append(problems, validateEntry(KindField, f.ID, f.Name, seen)...)
		switch f.Type {
		case FieldText, FieldLongText, FieldNumber, FieldDate, FieldOption, FieldDegree, FieldSubjectList:
		default:
			problems = append(problems, Problem{Code: CodeUnknownFieldType, Path: []Ref{{Kind: KindField, ID: f.ID}}, Value: string(f.Type)})
		}
		problems = append(problems, within(Ref{Kind: KindField, ID: f.ID}, validateOptions(f))...)
	}
	return problems
}

func validateDocuments(documents []Document, seen map[string]bool) []Problem {
	var problems []Problem
	for _, doc := range documents {
		problems = append(problems, validateEntry(KindDocument, doc.ID, doc.Name, seen)...)
		ref := Ref{Kind: KindDocument, ID: doc.ID}
		if len(doc.AcceptedTypes) == 0 {
			problems = append(problems, Problem{Code: CodeNoAcceptedTypes, Path: []Ref{ref}})
		}
		accepted := map[string]bool{}
		for _, mediaType := range doc.AcceptedTypes {
			switch {
			case !validMediaType(mediaType):
				problems = append(problems, Problem{Code: CodeInvalidAcceptedType, Path: []Ref{ref}, Value: mediaType})
			case accepted[mediaType]:
				problems = append(problems, Problem{Code: CodeRepeatedAcceptedType, Path: []Ref{ref}, Value: mediaType})
			}
			accepted[mediaType] = true
		}
		if doc.MaxSize <= 0 {
			problems = append(problems, Problem{Code: CodeNoMaxSize, Path: []Ref{ref}})
		}
	}
	return problems
}

// validMediaType accepts a lowercase type and subtype, without parameters or
// wildcards: "*/*" would let any file through the upload check.
func validMediaType(value string) bool {
	parsed, _, err := mime.ParseMediaType(value)
	return err == nil && parsed == value && strings.Contains(value, "/") && !strings.Contains(value, "*")
}

func validateOptions(f Field) []Problem {
	if f.Type != FieldOption {
		if len(f.Options) > 0 {
			return []Problem{{Code: CodeOptionsOutsideOption}}
		}
		return nil
	}
	if len(f.Options) == 0 {
		return []Problem{{Code: CodeOptionFieldNoOptions}}
	}
	var problems []Problem
	seenIDs := map[string]bool{}
	seenNames := map[string]bool{}
	for _, o := range f.Options {
		problems = append(problems, validateEntry(KindOption, o.ID, o.Name, seenIDs)...)
		problems = append(problems, repeatedName(KindOption, o.Name, seenNames)...)
	}
	return problems
}
