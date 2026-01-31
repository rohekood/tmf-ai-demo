package domain

// EvaluateEligibility determines the qualification status based on GIS polygon check and port availability.
func EvaluateEligibility(inPolygon bool, ports int) QualificationStatus {
	if !inPolygon {
		return StatusUnqualified
	}
	if ports <= 0 {
		return StatusUnqualified
	}
	return StatusQualified
}
