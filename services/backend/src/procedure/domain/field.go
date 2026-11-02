package domain

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

// An option is stored by its id, so that correcting its name later does not
// change what earlier answers mean.
type Option struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Document struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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
		if seenNames[o.Name] {
			problems = append(problems, Problem{Code: CodeRepeatedName, Path: []Ref{{Kind: KindOption}}, Value: o.Name})
		}
		seenNames[o.Name] = true
	}
	return problems
}
