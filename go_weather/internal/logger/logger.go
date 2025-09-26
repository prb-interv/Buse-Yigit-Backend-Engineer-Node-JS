package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog.Logger with additional context methods
type Logger struct {
	zerolog.Logger
}

// New creates a new structured logger instance
func New() *Logger {
	// Configure zerolog for beautiful console output (similar to Pino pretty)
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}

	logger := zerolog.New(output).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{Logger: logger}
}

// NewProduction creates a JSON logger for production (similar to Pino structured output)
func NewProduction() *Logger {
	logger := zerolog.New(os.Stdout).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{Logger: logger}
}

// Global logger instance
var globalLogger *Logger

func init() {
	globalLogger = New()
}

// Get returns the global logger instance
func Get() *Logger {
	return globalLogger
}

// SetGlobal sets the global logger instance
func SetGlobal(l *Logger) {
	globalLogger = l
}

// Weather request logging methods (similar to Pino structured logging)
func (l *Logger) WeatherRequest(location string, userID int) *zerolog.Event {
	return l.Info().
		Str("component", "weather").
		Str("action", "request").
		Str("location", location).
		Int("user_id", userID)
}

func (l *Logger) WeatherCompleted(location string, userID int, responseTime time.Duration, temperature float64, requestCount int) {
	l.Info().
		Str("component", "weather").
		Str("action", "completed").
		Str("location", location).
		Int("user_id", userID).
		Dur("response_time", responseTime).
		Float64("temperature", temperature).
		Int("request_count", requestCount).
		Msg("Weather request completed")
}

func (l *Logger) WeatherError(location string, userID int, err error, responseTime time.Duration) {
	l.Error().
		Str("component", "weather").
		Str("action", "error").
		Str("location", location).
		Int("user_id", userID).
		Dur("response_time", responseTime).
		Err(err).
		Msg("Weather request failed")
}

// API client logging methods
func (l *Logger) APIRequest(service, location, url string) *zerolog.Event {
	return l.Debug().
		Str("component", "api_client").
		Str("action", "request").
		Str("service", service).
		Str("location", location).
		Str("url", url)
}

func (l *Logger) APIResponse(service, location string, statusCode int, responseTime time.Duration) {
	l.Debug().
		Str("component", "api_client").
		Str("action", "response").
		Str("service", service).
		Str("location", location).
		Int("status_code", statusCode).
		Dur("response_time", responseTime).
		Msg("API response received")
}

func (l *Logger) APIError(service, location string, err error, responseTime time.Duration) {
	l.Error().
		Str("component", "api_client").
		Str("action", "error").
		Str("service", service).
		Str("location", location).
		Dur("response_time", responseTime).
		Err(err).
		Msg("API request failed")
}

// Aggregation logging methods
func (l *Logger) AggregationGroupCreated(location string) {
	l.Debug().
		Str("component", "aggregation").
		Str("action", "group_created").
		Str("location", location).
		Msg("Aggregation group created")
}

func (l *Logger) AggregationRequestAdded(location string, requestCount int, maxRequests int) {
	l.Debug().
		Str("component", "aggregation").
		Str("action", "request_added").
		Str("location", location).
		Int("current_requests", requestCount).
		Int("max_requests", maxRequests).
		Msg("Request added to aggregation group")
}

func (l *Logger) AggregationMaxReached(location string, requestCount int) {
	l.Info().
		Str("component", "aggregation").
		Str("action", "max_reached").
		Str("location", location).
		Int("request_count", requestCount).
		Msg("Max request limit reached, processing immediately")
}

func (l *Logger) AggregationTimerStarted(location string, waitTime time.Duration) {
	l.Debug().
		Str("component", "aggregation").
		Str("action", "timer_started").
		Str("location", location).
		Dur("wait_time", waitTime).
		Msg("Aggregation timer started")
}

func (l *Logger) AggregationProcessing(location string, requestCount int) {
	l.Info().
		Str("component", "aggregation").
		Str("action", "processing").
		Str("location", location).
		Int("request_count", requestCount).
		Msg("Processing aggregated requests")
}

// Database logging methods
func (l *Logger) DatabaseSave(location string, service1Temp, service2Temp float64, requestCount int) {
	l.Debug().
		Str("component", "database").
		Str("action", "save").
		Str("location", location).
		Float64("service1_temp", service1Temp).
		Float64("service2_temp", service2Temp).
		Int("request_count", requestCount).
		Msg("Weather data saved to database")
}

func (l *Logger) DatabaseError(operation string, err error) {
	l.Error().
		Str("component", "database").
		Str("action", "error").
		Str("operation", operation).
		Err(err).
		Msg("Database operation failed")
}

// Server logging methods
func (l *Logger) ServerStarted(port string) {
	l.Info().
		Str("component", "server").
		Str("action", "started").
		Str("port", port).
		Msg("Server started successfully")
}

func (l *Logger) ServerShutdown() {
	l.Info().
		Str("component", "server").
		Str("action", "shutdown").
		Msg("Server shutting down")
}

// Convenience methods for backward compatibility
func Info() *zerolog.Event {
	return globalLogger.Info()
}

func Error() *zerolog.Event {
	return globalLogger.Error()
}

func Debug() *zerolog.Event {
	return globalLogger.Debug()
}

func Warn() *zerolog.Event {
	return globalLogger.Warn()
}

func Fatal() *zerolog.Event {
	return globalLogger.Fatal()
}
