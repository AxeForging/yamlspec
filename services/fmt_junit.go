package services

import (
	"encoding/xml"
	"fmt"

	"github.com/AxeForging/yamlspec/domain"
)

// JUnitFormatter outputs results as JUnit XML for CI/CD integration
type JUnitFormatter struct{}

// NewJUnitFormatter creates a new JUnitFormatter
func NewJUnitFormatter() *JUnitFormatter {
	return &JUnitFormatter{}
}

type junitTestSuites struct {
	XMLName xml.Name         `xml:"testsuites"`
	Suites  []junitTestSuite `xml:"testsuite"`
}

type junitTestSuite struct {
	Name     string          `xml:"name,attr"`
	Tests    int             `xml:"tests,attr"`
	Failures int             `xml:"failures,attr"`
	Errors   int             `xml:"errors,attr"`
	Time     string          `xml:"time,attr"`
	Cases    []junitTestCase `xml:"testcase"`
}

type junitTestCase struct {
	Name      string        `xml:"name,attr"`
	ClassName string        `xml:"classname,attr"`
	Time      string        `xml:"time,attr"`
	Failure   *junitFailure `xml:"failure,omitempty"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Body    string `xml:",chardata"`
}

func (f *JUnitFormatter) Format(result *domain.SuiteResult) ([]byte, error) {
	suites := junitTestSuites{}

	for _, spec := range result.Specs {
		suite := junitTestSuite{
			Name: spec.Name,
			Time: fmt.Sprintf("%.3f", spec.Duration.Seconds()),
		}

		for _, desc := range spec.Describes {
			for _, ar := range desc.Assertions {
				suite.Tests++
				tc := junitTestCase{
					Name:      ar.Should,
					ClassName: fmt.Sprintf("%s.%s", spec.Name, desc.Name),
					Time:      "0.000",
				}

				if ar.Status == domain.StatusFailed || ar.Status == domain.StatusError {
					suite.Failures++
					tc.Failure = &junitFailure{
						Message: ar.Error,
						Type:    string(ar.Status),
						Body:    ar.Error,
					}
				}

				suite.Cases = append(suite.Cases, tc)
			}
		}

		if spec.Error != "" {
			suite.Tests++
			suite.Errors++
			suite.Cases = append(suite.Cases, junitTestCase{
				Name:      "spec execution",
				ClassName: spec.Name,
				Failure: &junitFailure{
					Message: spec.Error,
					Type:    "error",
					Body:    spec.Error,
				},
			})
		}

		suites.Suites = append(suites.Suites, suite)
	}

	output, err := xml.MarshalIndent(suites, "", "  ")
	if err != nil {
		return nil, err
	}

	return append([]byte(xml.Header), output...), nil
}
