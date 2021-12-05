package router

type HealthData struct {
	Data DataBlock `json:"data" binding:"required"`
}

type DataBlock struct {
	Values string `json:"values"`
	Dates  string `json:"dates"`
}
