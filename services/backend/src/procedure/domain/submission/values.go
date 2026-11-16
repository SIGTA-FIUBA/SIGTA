package submission

import (
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"backend/src/procedure/domain/definition"
)

const (
	// MaxText and MaxLongText bound what one field keeps. The numbers are a
	// guess, to confirm with the Secretaría.
	MaxText     = 200
	MaxLongText = 5000
	// MaxItems bounds a list of subjects.
	MaxItems = 100
	// dateLayout is the ISO 8601 date, as a browser date input sends it.
	dateLayout = "2006-01-02"
)

// number is a decimal written plainly, with a point: no exponent, no NaN and
// no infinity, which strconv would take.
var number = regexp.MustCompile(`^-?[0-9]+(\.[0-9]+)?$`)

// Value is what the applicant enters for a field: Items for a list of
// subjects, Text for every other type.
type Value struct {
	Text  string
	Items []string
}

// CheckValues returns the values as they are kept, without the spaces around
// them, or the problems that keep the submission from being accepted. A value
// that is sent must be valid even if its field is optional: leaving a field
// out is not sending it.
func CheckValues(def definition.ProcedureTypeDefinition, values map[string]Value) (map[string]Value, []Problem) {
	fields, _ := Requirements(def)
	asked := make(map[string]bool, len(fields))
	var problems []Problem
	clean := make(map[string]Value, len(values))
	for _, r := range fields {
		asked[r.Field.ID] = true
		v, sent := values[r.Field.ID]
		if !sent {
			if r.Required {
				problems = append(problems, FieldProblem(CodeMissing, r.Field.ID))
			}
			continue
		}
		cleaned, ok := checkValue(r.Field, v)
		if !ok {
			problems = append(problems, FieldProblem(CodeInvalid, r.Field.ID))
			continue
		}
		clean[r.Field.ID] = cleaned
	}
	var notAsked []string
	for id := range values {
		if !asked[id] {
			notAsked = append(notAsked, id)
		}
	}
	slices.Sort(notAsked)
	for _, id := range notAsked {
		problems = append(problems, FieldProblem(CodeNotAsked, id))
	}
	if len(problems) > 0 {
		return nil, problems
	}
	return clean, nil
}

func checkValue(f definition.Field, v Value) (Value, bool) {
	if f.Type == definition.FieldSubjectList {
		if v.Text != "" || len(v.Items) == 0 || len(v.Items) > MaxItems {
			return Value{}, false
		}
		items := make([]string, len(v.Items))
		for i, item := range v.Items {
			items[i] = strings.TrimSpace(item)
			if !validText(items[i], MaxText) {
				return Value{}, false
			}
		}
		return Value{Items: items}, true
	}
	if len(v.Items) > 0 {
		return Value{}, false
	}
	text := strings.TrimSpace(v.Text)
	var ok bool
	switch f.Type {
	case definition.FieldText, definition.FieldDegree:
		ok = validText(text, MaxText)
	case definition.FieldLongText:
		ok = validText(text, MaxLongText)
	case definition.FieldNumber:
		ok = validText(text, MaxText) && number.MatchString(text)
	case definition.FieldDate:
		_, err := time.Parse(dateLayout, text)
		ok = err == nil
	case definition.FieldOption:
		ok = slices.ContainsFunc(f.Options, func(o definition.Option) bool { return o.ID == text })
	}
	return Value{Text: text}, ok
}

func validText(text string, maxRunes int) bool {
	return text != "" && utf8.ValidString(text) && utf8.RuneCountInString(text) <= maxRunes
}
