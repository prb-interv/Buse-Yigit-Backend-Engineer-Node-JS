package services

import (
	"fmt"
	"sync"
	"time"

	"goweather/internal/clients"
	"goweather/internal/config"
	"goweather/internal/database"
	"goweather/internal/logger"
	"goweather/pkg/types"
)

type WeatherService struct {
	weatherAPIClient  *clients.WeatherAPIClient
	weatherStackClient *clients.WeatherStackClient
	database          *database.Database
	logger            *logger.Logger
	
	aggregationMap    map[string]*AggregationGroup
	aggregationMutex  sync.RWMutex
	
	maxRequests       int
	waitTime          time.Duration
}

type AggregationGroup struct {
	Location     string
	Requests     []types.AggregationRequest
	Timer        *time.Timer
	Mutex        sync.Mutex
	MaxRequests  int
	WaitTime     time.Duration
	IsProcessing bool
}

func NewWeatherService(db *database.Database, cfg *config.Config) *WeatherService {
	return &WeatherService{
		weatherAPIClient:  clients.NewWeatherAPIClient(cfg.WeatherAPIKey, cfg.APITimeout),
		weatherStackClient: clients.NewWeatherStackClient(cfg.WeatherStackKey, cfg.APITimeout),
		database:          db,
		logger:            logger.Get(),
		aggregationMap:    make(map[string]*AggregationGroup),
		maxRequests:       cfg.MaxRequests,
		waitTime:          cfg.WaitTime,
	}
}

func (s *WeatherService) GetWeather(location string) (*types.WeatherResponse, error) {
	group := s.getOrCreateAggregationGroup(location)

	responseChan := make(chan types.WeatherResponse, 1)
	errorChan := make(chan error, 1)
	
	request := types.AggregationRequest{
		Location: location,
		Response: responseChan,
		Error:    errorChan,
	}	
	group.Mutex.Lock()
	
	// Eğer group processing durumundaysa, yeni request ekleme
	if group.IsProcessing {
		// Processing durumundaki gruplara yeni request eklemiyoruz
		//  processing sırasında gelenler aynı grup içinde birikir ve timer ile yeni batch açılır.
		startTimer := (group.Timer == nil)
		group.Requests = append(group.Requests, request)
		if startTimer {
			group.Timer = time.AfterFunc(group.WaitTime, func() {
				group.Mutex.Lock()
				batch, ok := s.triggerLocked(group)
				group.Mutex.Unlock()
				if !ok {
					return
				}
				s.processAggregationGroupWithBatch(group, batch)
			})
		}
		group.Mutex.Unlock()
		return s.waitForResponse(responseChan, errorChan)
	}
	
	group.Requests = append(group.Requests, request)
	requestCount := len(group.Requests)
	isFirstRequest := (requestCount == 1)
	
	// Max request limitine ulaşıldığında hemen işle
	if requestCount >= group.MaxRequests {
		s.logger.AggregationMaxReached(location, requestCount)
		if group.Timer != nil {
			group.Timer.Stop()
			group.Timer = nil
		}
		batch, ok := s.triggerLocked(group)
		group.Mutex.Unlock()
		if ok {
			go s.processAggregationGroupWithBatch(group, batch)
		}
		return s.waitForResponse(responseChan, errorChan)
	}
	
	// İlk request ise timer başlat
	if isFirstRequest {
		group.Timer = time.AfterFunc(group.WaitTime, func() {
			group.Mutex.Lock()
			batch, ok := s.triggerLocked(group)
			group.Mutex.Unlock()
			if !ok {
				return
			}
			s.processAggregationGroupWithBatch(group, batch)
		})
	}
	
	group.Mutex.Unlock()
	return s.waitForResponse(responseChan, errorChan)
}

// handleNewRequestImmediately handles requests when the current group is processing
func (s *WeatherService) handleNewRequestImmediately(location string, responseChan chan types.WeatherResponse, errorChan chan error) (*types.WeatherResponse, error) {
	s.logger.Info().
		Str("component", "aggregation").
		Str("action", "immediate_processing").
		Str("location", location).
		Msg("Processing request immediately")
	
	// Fetch weather data directly without aggregation
	weatherData, err := s.fetchWeatherData(location, 1)
	if err != nil {
		s.logger.Error().
			Str("component", "aggregation").
			Str("action", "immediate_processing_error").
			Str("location", location).
			Err(err).
			Msg("Weather data not fetched in immediate processing")
		return nil, err
	}
	
	response := types.WeatherResponse{
		Location:    weatherData.Location,
		Temperature: weatherData.AverageTemp,
	}
	
	s.logger.Info().
		Str("component", "aggregation").
		Str("action", "immediate_processing_completed").
		Str("location", location).
		Float64("temperature", weatherData.AverageTemp).
		Msg("Immediate processing completed")
	return &response, nil
}

func (s *WeatherService) getOrCreateAggregationGroup(location string) *AggregationGroup {
	s.aggregationMutex.Lock()
	defer s.aggregationMutex.Unlock()
	
	group, exists := s.aggregationMap[location]
	if !exists {
		group = &AggregationGroup{
			Location:     location,
			Requests:     make([]types.AggregationRequest, 0),
			MaxRequests:  s.maxRequests,
			WaitTime:     s.waitTime,
			IsProcessing: false,
		}
		s.aggregationMap[location] = group
		s.logger.AggregationGroupCreated(location)
	}
	
	return group
}

func (s *WeatherService) processAggregationGroup(group *AggregationGroup) {
	// stop timer outside of mutex
	if group.Timer != nil {
		group.Timer.Stop()
		group.Timer = nil
	}
	
	// Request'leri kopyala ve mutex'i serbest bırak
	group.Mutex.Lock()
	// IsProcessing zaten set edildi
	requests := make([]types.AggregationRequest, len(group.Requests))
	copy(requests, group.Requests)
	requestCount := len(requests)
	group.Requests = nil
	group.Mutex.Unlock()
	
	s.logger.AggregationProcessing(group.Location, requestCount)
	
	weatherData, err := s.fetchWeatherData(group.Location, requestCount)
	if err != nil {
		s.logger.Error().
			Str("component", "aggregation").
			Str("action", "fetch_weather_error").
			Str("location", group.Location).
			Int("request_count", requestCount).
			Err(err).
			Msg("Weather data not fetched")
		
		for _, req := range requests {
			req.Error <- err
		}
		
		// Clean up the group after error
		group.Mutex.Lock()
		group.IsProcessing = false
		// Eğer bekleyen istekler varsa ve timer yoksa yeni batch için timer başlat
		if len(group.Requests) > 0 && group.Timer == nil {
			group.Timer = time.AfterFunc(group.WaitTime, func() {
				group.Mutex.Lock()
				batch, ok := s.triggerLocked(group)
				group.Mutex.Unlock()
				if !ok {
					return
				}
				s.processAggregationGroupWithBatch(group, batch)
			})
		}
		group.Mutex.Unlock()
		return
	}
	
	response := types.WeatherResponse{
		Location:    weatherData.Location,
		Temperature: weatherData.AverageTemp,
	}
	
	for _, req := range requests {
		req.Response <- response
	}
	
	s.logger.Info().
		Str("component", "aggregation").
		Str("action", "completed").
		Str("location", group.Location).
		Float64("temperature", weatherData.AverageTemp).
		Int("request_count", requestCount).
		Msg("Aggregation completed")
	
	// Clean up the group after successful processing
	group.Mutex.Lock()
	group.IsProcessing = false
	// Eğer bekleyen istekler varsa ve timer yoksa yeni batch için timer başlasın
	if len(group.Requests) > 0 && group.Timer == nil {
		group.Timer = time.AfterFunc(group.WaitTime, func() {
			group.Mutex.Lock()
			batch, ok := s.triggerLocked(group)
			group.Mutex.Unlock()
			if !ok {
				return
			}
			s.processAggregationGroupWithBatch(group, batch)
		})
	}
	group.Mutex.Unlock()
}

// cleanupAggregationGroup removes the group from the map and resets its state
func (s *WeatherService) cleanupAggregationGroup(group *AggregationGroup) {
	group.Mutex.Lock()
	group.Requests = make([]types.AggregationRequest, 0)
	group.Timer = nil
	group.IsProcessing = false
	group.Mutex.Unlock()
	
	// Remove from aggregation map
	s.aggregationMutex.Lock()
	delete(s.aggregationMap, group.Location)
	s.aggregationMutex.Unlock()
	
	s.logger.Debug().
		Str("component", "aggregation").
		Str("action", "group_cleaned").
		Str("location", group.Location).
		Msg("Aggregation group cleaned up")
}

// fetch data 
func (s *WeatherService) fetchWeatherData(location string, requestCount int) (*types.WeatherData, error) {
	
	var service1Temp, service2Temp float64
	var service1Err, service2Err error
	
	var wg sync.WaitGroup
	wg.Add(2)
	
	// weatherapi
	go func() {
		defer wg.Done()
		service1Temp, service1Err = s.weatherAPIClient.GetTemperature(location)
	}()
	
	// weather stack
	go func() {
		defer wg.Done()
		service2Temp, service2Err = s.weatherStackClient.GetTemperature(location)
	}()
	
	wg.Wait()
	
	if service1Err != nil {
		return nil, fmt.Errorf("WeatherAPI.com hatası: %v", service1Err)
	}
	if service2Err != nil {
		return nil, fmt.Errorf("WeatherStack.com hatası: %v", service2Err)
	}
	averageTemp := (service1Temp + service2Temp) / 2
	
	weatherData := &types.WeatherData{
		Location:     location,
		Service1Temp: service1Temp,
		Service2Temp: service2Temp,
		AverageTemp:  averageTemp,
		RequestCount: requestCount,
	}
	
	// async save to database
	go func() {
		query := &types.WeatherQuery{
			Location:     location,
			Service1Temp: service1Temp,
			Service2Temp: service2Temp,
			RequestCount: requestCount,
		}
		
		if err := s.database.SaveWeatherQuery(query); err != nil {
			s.logger.DatabaseError("save_weather_query", err)
		} else {
			s.logger.DatabaseSave(location, service1Temp, service2Temp, requestCount)
		}
	}()
	
	return weatherData, nil
}


func (s *WeatherService) waitForResponse(responseChan chan types.WeatherResponse, errorChan chan error) (*types.WeatherResponse, error) {
	select {
	case response := <-responseChan:
		return &response, nil
	case err := <-errorChan:
		return nil, err
	}
}

// ——— yardımcı fonksiyonlar (yorum eklemeden) ———

func (s *WeatherService) triggerLocked(group *AggregationGroup) ([]types.AggregationRequest, bool) {
	if group.Timer != nil {
		group.Timer.Stop()
		group.Timer = nil
	}
	if group.IsProcessing {
		return nil, false
	}
	group.IsProcessing = true
	batch := make([]types.AggregationRequest, len(group.Requests))
	copy(batch, group.Requests)
	group.Requests = nil
	return batch, true
}

func (s *WeatherService) processAggregationGroupWithBatch(group *AggregationGroup, batch []types.AggregationRequest) {
	requestCount := len(batch)
	s.logger.AggregationProcessing(group.Location, requestCount)

	weatherData, err := s.fetchWeatherData(group.Location, requestCount)
	if err != nil {
		s.logger.Error().
			Str("component", "aggregation").
			Str("action", "fetch_weather_error_batch").
			Str("location", group.Location).
			Int("request_count", requestCount).
			Err(err).
			Msg("Weather data not fetched in batch processing")
		for _, req := range batch {
			req.Error <- err
		}
		group.Mutex.Lock()
		group.IsProcessing = false
		if len(group.Requests) > 0 && group.Timer == nil {
			group.Timer = time.AfterFunc(group.WaitTime, func() {
				group.Mutex.Lock()
				next, ok := s.triggerLocked(group)
				group.Mutex.Unlock()
				if !ok {
					return
				}
				s.processAggregationGroupWithBatch(group, next)
			})
			s.logger.AggregationTimerStarted(group.Location, group.WaitTime)
		}
		group.Mutex.Unlock()
		return
	}

	response := types.WeatherResponse{
		Location:    weatherData.Location,
		Temperature: weatherData.AverageTemp,
	}
	for _, req := range batch {
		req.Response <- response
	}

	s.logger.Info().
		Str("component", "aggregation").
		Str("action", "batch_completed").
		Str("location", group.Location).
		Float64("temperature", weatherData.AverageTemp).
		Int("request_count", requestCount).
		Msg("Batch processing completed")

	group.Mutex.Lock()
	group.IsProcessing = false
	if len(group.Requests) > 0 && group.Timer == nil {
		group.Timer = time.AfterFunc(group.WaitTime, func() {
			group.Mutex.Lock()
			next, ok := s.triggerLocked(group)
			group.Mutex.Unlock()
			if !ok {
				return
			}
			s.processAggregationGroupWithBatch(group, next)
		})
		s.logger.AggregationTimerStarted(group.Location, group.WaitTime)
	}
	group.Mutex.Unlock()
}
