package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"goweather/internal/logger"
	"goweather/internal/services"
)

type WeatherHandler struct {
	weatherService *services.WeatherService
	logger         *logger.Logger
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type WeatherResponse struct {
	Location    string  `json:"location"`
	Temperature float64 `json:"temperature"`
}

func NewWeatherHandler(weatherService *services.WeatherService) *WeatherHandler {
	return &WeatherHandler{
		weatherService: weatherService,
		logger:         logger.Get(),
	}
}

func (h *WeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	// Generate a simple user ID for demo purposes (in real app, this would come from auth)
	userID := 123
	
	// Query parameter kontrolü
	location := r.URL.Query().Get("q")
	if location == "" {
		h.logger.WeatherError(location, userID, nil, time.Since(startTime))
		h.sendError(w, http.StatusBadRequest, "MISSING_LOCATION", "Location parameter 'q' is required")
		return
	}

	// Log the weather request (similar to Pino example)
	h.logger.WeatherRequest(location, userID).Msg("User requested weather")

	// Weather service çağrısı
	weatherResp, err := h.weatherService.GetWeather(location)
	responseTime := time.Since(startTime)
	
	if err != nil {
		h.logger.WeatherError(location, userID, err, responseTime)
		h.sendError(w, http.StatusInternalServerError, "WEATHER_SERVICE_ERROR", "Failed to fetch weather data")
		return
	}

	// Başarılı response - structured logging like Pino
	h.logger.WeatherCompleted(location, userID, responseTime, weatherResp.Temperature, 1)

	response := WeatherResponse{
		Location:    weatherResp.Location,
		Temperature: weatherResp.Temperature,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error().
			Str("component", "handler").
			Str("action", "json_encode_error").
			Err(err).
			Msg("JSON encoding failed")
	}
}

func (h *WeatherHandler) sendError(w http.ResponseWriter, statusCode int, errorCode, message string) {
	errorResp := ErrorResponse{
		Error:   errorCode,
		Code:    statusCode,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		h.logger.Error().
			Str("component", "handler").
			Str("action", "error_encode_json").
			Str("error_code", errorCode).
			Int("status_code", statusCode).
			Err(err).
			Msg("Error encoding JSON response")
	}
}
