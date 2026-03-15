package scanfailure

import "github.com/armosec/armoapi-go/scanfailure"

// Re-export types from armoapi-go for backward compatibility.
// New code should import from github.com/armosec/armoapi-go/scanfailure directly.

type ScanFailureCase = scanfailure.ScanFailureCase
type ScanFailureReport = scanfailure.ScanFailureReport

const (
	ScanFailureUnknown        = scanfailure.ScanFailureUnknown
	ScanFailureCVE            = scanfailure.ScanFailureCVE
	ScanFailureSBOMGeneration = scanfailure.ScanFailureSBOMGeneration
	ScanFailureOOMKilled      = scanfailure.ScanFailureOOMKilled
	ScanFailureBackendPost    = scanfailure.ScanFailureBackendPost
)
