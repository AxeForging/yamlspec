package services

import (
	"encoding/json"

	"github.com/AxeForging/yamlspec/domain"
)

// JSONFormatter outputs results as pretty-printed JSON
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSONFormatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (f *JSONFormatter) Format(result *domain.SuiteResult) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}
