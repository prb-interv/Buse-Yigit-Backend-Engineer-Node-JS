package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"goweather/internal/config"
	"goweather/internal/database"
	"goweather/internal/handlers"
	"goweather/internal/logger"
	"goweather/internal/services"
	"goweather/pkg/types"
)

func main() {
	// Initialize logger
	log := logger.Get()
	
	cfg := config.LoadConfig()
	
	log.Info().
		Str("component", "server").
		Str("action", "startup").
		Str("port", cfg.ServerPort).
		Str("database_path", cfg.DatabasePath).
		Int("max_requests", cfg.MaxRequests).
		Bool("debug_mode", cfg.DebugMode).
		Msg("Starting weather API server")
	
	db, err := database.NewDatabase(cfg.DatabasePath)
	if err != nil {
		log.Fatal().
			Str("component", "server").
			Str("action", "database_connection_failed").
			Err(err).
			Msg("Database connection failed")
	}
	defer db.Close()
	
	
	weatherService := services.NewWeatherService(db, cfg)
	weatherHandler := handlers.NewWeatherHandler(weatherService)

	log.Debug().
		Str("component", "server").
		Str("action", "config_loaded").
		Str("port", cfg.ServerPort).
		Str("database_path", cfg.DatabasePath).
		Int("max_requests", cfg.MaxRequests).
		Msg("Configuration loaded")
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Weather API Server is running", "status": "ok"}`)
	})
	
	// Debug mode'da queries endpoint'i ekle
	if cfg.DebugMode {
		http.HandleFunc("/queries", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			
			queries, err := db.GetWeatherQueries()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				errorResp := types.ErrorResponse{
					Error:   "Data not found",
					Message: err.Error(),
				}
				json.NewEncoder(w).Encode(errorResp)
				return
			}
			
			json.NewEncoder(w).Encode(queries)
		})
	}
	
	http.HandleFunc("/weather", weatherHandler.GetWeather)
	
	
	port := ":" + cfg.ServerPort
	log.ServerStarted(cfg.ServerPort)
	log.Info().
		Str("component", "server").
		Str("action", "ready").
		Str("test_url", fmt.Sprintf("http://localhost%s/weather?q=Istanbul", port)).
		Msg("Server ready to accept requests")
	
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal().
			Str("component", "server").
			Str("action", "server_start_failed").
			Err(err).
			Msg("Server failed to start")
	}
}