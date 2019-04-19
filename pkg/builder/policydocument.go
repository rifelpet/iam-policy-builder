package builder

import (
	"fmt"

	"github.com/rifelpet/iam-policy-builder/pkg/iam"
)

func BuildDocument(usages []*iam.AWSServiceUsage) (*iam.PolicyDocument, error) {
	document := &iam.PolicyDocument{
		Version:   "2012-10-17",
		Statement: make([]*iam.StatementEntry, 0),
	}
	services := make(map[string]map[string]bool)
	for _, usage := range usages {
		svc := usage.Service
		if _, ok := services[svc]; !ok {
			services[svc] = make(map[string]bool)
		}
		for _, action := range usage.FunctionCalls {
			services[svc][action] = true
		}
	}

	for svc, actions := range services {
		statement := &iam.StatementEntry{
			Sid:      svc,
			Effect:   "allow",
			Action:   make([]string, 0),
			Resource: []string{"*"},
		}
		for action := range actions {
			statement.Action = append(statement.Action, fmt.Sprintf("%v:%v", svc, action))
		}
		document.Statement = append(document.Statement, statement)
	}

	return document, nil
}
