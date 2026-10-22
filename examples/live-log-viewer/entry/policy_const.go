package entry

import "time"

// Default policy numbers, named so the mnd gate does not flag them.
const (
	policyBatchRecords = 256
	policyBatchBytes   = 256 << 10
	policyMaxRecord    = 64 << 10
	policyPoll         = 250 * time.Millisecond
	policyMultiLines   = 200
	policyMultiBytes   = 256 << 10
	policyMultiHold    = 500 * time.Millisecond
	policyQueryTimeout = 5 * time.Second
	policyPruneBatch   = 1000
	policyPageLimit    = 200
)
