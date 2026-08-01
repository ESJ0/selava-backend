package validator

import "strings"

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []FieldError

func (v ValidationErrors) Error() string {
	msgs := make([]string, len(v))
	for i, fe := range v {
		msgs[i] = fe.Field + ": " + fe.Message
	}
	return strings.Join(msgs, "; ")
}

func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}
