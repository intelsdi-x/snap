package ipmi

// Abstract type for ipmi backend.
type IpmiAL interface {
	BatchExecRaw(requests []IpmiRequest, nSim int) ([]IpmiResponse, error)
}

// Defines request parameter passed to abstraction layer.
type IpmiRequest struct {
	Data    []byte
	Channel int16
	Slave   uint8
}

// Defines response data.
type IpmiResponse struct {
	Data []byte
}

// Vendor exposed structure. Defines request content and response format.
// List of submetrics for given format should be concatenated with MetricsRoot
// to specify full metric name.
type RequestDescription struct {
	Request     IpmiRequest
	MetricsRoot string
	Format      ParserFormat
}

// Defines interface that all response formats must implement.
// GetMetrics() should return all available submetrics for given format.
// Main metric value should have label "" (empty string).
// Validate() should check response correctness. Nil is returned when response
// is correct.
// Parse() extracts submetrics from binary data.
type ParserFormat interface {
	GetMetrics() []string
	Validate(response []byte) error
	Parse(response []byte) map[string]uint16
}
