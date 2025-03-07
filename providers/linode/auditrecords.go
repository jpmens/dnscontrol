package linode

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	a.Add("CAA", rejectif.CaaFlagIsNonZero) // Last verified 2022-03-25

	a.Add("CAA", rejectif.CaaTargetContainsWhitespace) // Last verified 2023-01-15

	return a.Audit(records)
}
