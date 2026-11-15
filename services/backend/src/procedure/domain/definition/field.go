package definition

import (
	"mime"
	"slices"
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
	MaxSizeBytes  int64    `json:"max_size,omitempty"`
}

func validateFields(fields []Field, seen map[string]bool) []ValidationProblem {
	var problems []ValidationProblem
	for _, f := range fields {
		problems = append(problems, validateEntry(KindField, f.ID, f.Name, seen)...)
		switch f.Type {
		case FieldText, FieldLongText, FieldNumber, FieldDate, FieldOption, FieldDegree, FieldSubjectList:
		default:
			problems = append(problems, ValidationProblem{Code: CodeUnknownFieldType, Path: []Element{{Kind: KindField, ID: f.ID}}, Value: string(f.Type)})
		}
		problems = append(problems, within(Element{Kind: KindField, ID: f.ID}, validateOptions(f))...)
	}
	return problems
}

func validateDocuments(documents []Document, seen map[string]bool) []ValidationProblem {
	var problems []ValidationProblem
	for _, doc := range documents {
		problems = append(problems, validateEntry(KindDocument, doc.ID, doc.Name, seen)...)
		ref := Element{Kind: KindDocument, ID: doc.ID}
		if len(doc.AcceptedTypes) == 0 {
			problems = append(problems, ValidationProblem{Code: CodeNoAcceptedTypes, Path: []Element{ref}})
		}
		accepted := map[string]bool{}
		for _, mediaType := range doc.AcceptedTypes {
			switch {
			case !validMediaType(mediaType):
				problems = append(problems, ValidationProblem{Code: CodeInvalidAcceptedType, Path: []Element{ref}, Value: mediaType})
			case !slices.Contains(DetectedTypes, mediaType):
				problems = append(problems, ValidationProblem{Code: CodeUndetectedAcceptedType, Path: []Element{ref}, Value: mediaType})
			case accepted[mediaType]:
				problems = append(problems, ValidationProblem{Code: CodeRepeatedAcceptedType, Path: []Element{ref}, Value: mediaType})
			}
			accepted[mediaType] = true
		}
		if doc.MaxSizeBytes <= 0 {
			problems = append(problems, ValidationProblem{Code: CodeNoMaxSize, Path: []Element{ref}})
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

// DetectedTypes are the types that the upload can recognize in a file, the
// only ones a document can accept: the upload checks the type it detects in
// the content, not the one the browser sends. They are the ones that
// http.DetectContentType returns, without parameters, except
// application/octet-stream, which it returns for any content it does not
// recognize. A Word or OpenDocument file is detected as application/zip, so
// those formats cannot be asked for yet.
var DetectedTypes = []string{
	"application/ogg",
	"application/pdf",
	"application/postscript",
	"application/vnd.ms-fontobject",
	"application/wasm",
	"application/x-gzip",
	"application/x-rar-compressed",
	"application/zip",
	"audio/aiff",
	"audio/midi",
	"audio/mpeg",
	"audio/wave",
	"font/collection",
	"font/otf",
	"font/ttf",
	"font/woff",
	"font/woff2",
	"image/bmp",
	"image/gif",
	"image/jpeg",
	"image/png",
	"image/vnd.microsoft.icon",
	"image/webp",
	"image/x-icon",
	"text/html",
	"text/plain",
	"text/xml",
	"video/avi",
	"video/mp4",
	"video/webm",
}

func validateOptions(f Field) []ValidationProblem {
	if f.Type != FieldOption {
		if len(f.Options) > 0 {
			return []ValidationProblem{{Code: CodeOptionsOutsideOption}}
		}
		return nil
	}
	if len(f.Options) == 0 {
		return []ValidationProblem{{Code: CodeOptionFieldNoOptions}}
	}
	var problems []ValidationProblem
	seenIDs := map[string]bool{}
	seenNames := map[string]bool{}
	for _, o := range f.Options {
		problems = append(problems, validateEntry(KindOption, o.ID, o.Name, seenIDs)...)
		problems = append(problems, repeatedName(KindOption, o.Name, seenNames)...)
	}
	return problems
}
