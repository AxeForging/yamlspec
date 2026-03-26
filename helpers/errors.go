package helpers

import "fmt"

// YamlspecError is a structured error with operation context
type YamlspecError struct {
	Operation string
	Details   string
	Err       error
}

func (e *YamlspecError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s: %v", e.Operation, e.Details, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Operation, e.Err)
}

func (e *YamlspecError) Unwrap() error {
	return e.Err
}

// WrapError creates a structured error
func WrapError(operation, details string, err error) error {
	return &YamlspecError{
		Operation: operation,
		Details:   details,
		Err:       err,
	}
}
