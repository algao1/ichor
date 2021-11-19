package router

type HealthData struct {
	Glucose GlucoseData `json:"glucose" binding:"required"`
}

type GlucoseData struct {
	Values string `json:"values"`
	Dates  string `json:"dates"`
}
