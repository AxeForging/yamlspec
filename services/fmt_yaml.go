package services

import (
	"github.com/AxeForging/yamlspec/domain"
	"gopkg.in/yaml.v3"
)

// YAMLFormatter outputs results as YAML
type YAMLFormatter struct{}

// NewYAMLFormatter creates a new YAMLFormatter
func NewYAMLFormatter() *YAMLFormatter {
	return &YAMLFormatter{}
}

func (f *YAMLFormatter) Format(result *domain.SuiteResult) ([]byte, error) {
	return yaml.Marshal(result)
}
