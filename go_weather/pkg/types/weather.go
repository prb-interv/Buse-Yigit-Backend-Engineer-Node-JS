package types

// WeatherRequest 
type WeatherRequest struct {
	Location string `json:"location" validate:"required"`
}

// WeatherResponse
type WeatherResponse struct {
	Location    string  `json:"location"`
	Temperature float64 `json:"temperature"`
}

// WeatherData Combined
type WeatherData struct {
	Location         string  `json:"location"`
	Service1Temp     float64 `json:"service_1_temperature"`
	Service2Temp     float64 `json:"service_2_temperature"`
	AverageTemp      float64 `json:"average_temperature"`
	RequestCount     int     `json:"request_count"`
}

// DB Query Schema
type WeatherQuery struct {
	ID                int     `json:"id" db:"id"`
	Location          string  `json:"location" db:"location"`
	Service1Temp      float64 `json:"service_1_temperature" db:"service_1_temperature"`
	Service2Temp      float64 `json:"service_2_temperature" db:"service_2_temperature"`
	RequestCount      int     `json:"request_count" db:"request_count"`
}

// WeatherAPIResponse
type WeatherAPIResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

// WeatherStackResponse
type WeatherStackResponse struct {
	Current struct {
		Temperature float64 `json:"temperature"`
	} `json:"current"`
}

// ErrorResponse
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// AggregationRequest
type AggregationRequest struct {
	Location  string
	Response  chan WeatherResponse
	Error     chan error
}
