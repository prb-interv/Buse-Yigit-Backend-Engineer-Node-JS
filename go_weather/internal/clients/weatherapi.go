package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	
	"goweather/internal/logger"
	"goweather/pkg/types"
)

// WeatherAPIClient 
type WeatherAPIClient struct {
	APIKey string
	BaseURL string
	Client  *http.Client
	logger  *logger.Logger
}

// NewWeatherAPIClient 
func NewWeatherAPIClient(apiKey string, timeout time.Duration) *WeatherAPIClient {
	return &WeatherAPIClient{
		APIKey:  apiKey,
		BaseURL: "http://api.weatherapi.com/v1/forecast.json",
		Client: &http.Client{
			Timeout: timeout, 
		},
		logger: logger.Get(),
	}
}

// GetWeather 
func (c *WeatherAPIClient) GetWeather(location string) (*types.WeatherAPIResponse, error) {
	startTime := time.Now()
	url := fmt.Sprintf("%s?key=%s&q=%s&days=1&aqi=no&alerts=no", 
		c.BaseURL, c.APIKey, location)

	c.logger.APIRequest("weatherapi", location, url).Msg("API request started")
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.logger.APIError("weatherapi", location, err, time.Since(startTime))
		return nil, fmt.Errorf("HTTP request creation failed: %v", err)
	}

	// send req
	resp, err := c.Client.Do(req)
	responseTime := time.Since(startTime)
	
	if err != nil {
		c.logger.APIError("weatherapi", location, err, responseTime)
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// Status code check
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		apiErr := fmt.Errorf("API error (Status: %d): %s", resp.StatusCode, string(body))
		c.logger.APIError("weatherapi", location, apiErr, responseTime)
		return nil, apiErr
	}
	
	c.logger.APIResponse("weatherapi", location, resp.StatusCode, responseTime)

	// Response body read
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Response read failed: %v", err)
	}

	// JSON parse
	var weatherResp types.WeatherAPIResponse
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		return nil, fmt.Errorf("JSON parse failed: %v", err)
	}

	return &weatherResp, nil
}

// GetTemperature 
func (c *WeatherAPIClient) GetTemperature(location string) (float64, error) {
	weather, err := c.GetWeather(location)
	if err != nil {
		return 0, err
	}

	return weather.Current.TempC, nil
}
