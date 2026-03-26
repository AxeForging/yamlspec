package services

import "github.com/AxeForging/yamlspec/domain"

// Formatter formats test results into a specific output format
type Formatter interface {
	Format(result *domain.SuiteResult) ([]byte, error)
}
