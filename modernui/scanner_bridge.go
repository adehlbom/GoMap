package modernui

// ScannerFunc defines the type for the network scanner function
type ScannerFunc func(subnetOverride string, portRangeStr string) (interface{}, error)

// Global variable to store the scanner function
var scannerImplementation ScannerFunc

// RegisterScanner registers the scanner implementation to be used by the UI
func RegisterScanner(scanner ScannerFunc) {
	scannerImplementation = scanner
}

// ExecuteScan runs the registered scanner with the given parameters
func ExecuteScan(subnet string, portRange string) (interface{}, error) {
	if scannerImplementation == nil {
		return nil, nil
	}
	return scannerImplementation(subnet, portRange)
}
