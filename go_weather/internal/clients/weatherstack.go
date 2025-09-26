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

// WeatherStackClient 
type WeatherStackClient struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
	logger  *logger.Logger
}

func NewWeatherStackClient(apiKey string, timeout time.Duration) *WeatherStackClient {
	return &WeatherStackClient{
		APIKey:  apiKey,
		BaseURL: "http://api.weatherstack.com/current", //HTTP
		Client: &http.Client{
			Timeout: timeout, 
		},
		logger: logger.Get(),
	}
}


func (c *WeatherStackClient) GetWeather(location string) (*types.WeatherStackResponse, error) {
	startTime := time.Now()
	url := fmt.Sprintf("%s?access_key=%s&query=%s", 
		c.BaseURL, c.APIKey, location)

	c.logger.APIRequest("weatherstack", location, url).Msg("API request started")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.logger.APIError("weatherstack", location, err, time.Since(startTime))
		return nil, fmt.Errorf("HTTP isteği oluşturulamadı: %v", err)
	}

	resp, err := c.Client.Do(req)
	responseTime := time.Since(startTime)
	
	if err != nil {
		c.logger.APIError("weatherstack", location, err, responseTime)
		return nil, fmt.Errorf("HTTP isteği başarısız: %v", err)
	}
	defer resp.Body.Close()

	//Status code check
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		apiErr := fmt.Errorf("API hatası (Status: %d): %s", resp.StatusCode, string(body))
		c.logger.APIError("weatherstack", location, apiErr, responseTime)
		return nil, apiErr
	}
	
	c.logger.APIResponse("weatherstack", location, resp.StatusCode, responseTime)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Response okunamadı: %v", err)
	}

	var weatherResp types.WeatherStackResponse
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		return nil, fmt.Errorf("JSON parse hatası: %v", err)
	}

	return &weatherResp, nil
}

// Get temperature 
func (c *WeatherStackClient) GetTemperature(location string) (float64, error) {
	weather, err := c.GetWeather(location)
	if err != nil {
		return 0, err
	}

	return weather.Current.Temperature, nil
}
