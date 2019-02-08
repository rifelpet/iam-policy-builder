package iam

// AWSServiceUsage represents the importing and usage of an AWS Service in one go file
type AWSServiceUsage struct {
	ImportName    string
	Service       string
	ClientNames   []string
	FunctionCalls []string // TODO: add support for parameters
}

// PolicyDocument represents an IAM Policy
type PolicyDocument struct {
	Version   string
	Statement []*StatementEntry
}

// StatementEntry represents an individual Statement in a policy document
type StatementEntry struct {
	Effect   string
	Action   []string
	Resource []string
}
