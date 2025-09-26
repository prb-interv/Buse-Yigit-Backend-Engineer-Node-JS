package main

/*
Weather API Test Script
İstanbul: 11 eşzamanlı istek
Ankara: 3 istek, 1 saniye aralıklarla

Test kapsamında İstanbul için ilk 10 isteğin anında ve eşzamanlı, diğer isteğin ise 5 saniye boyunca diğer istekleri dinledikten sonra cevap dönmesi beklenir.
Ankara için ise 3 istek 1 saniye aralıklarla beklenir. İlk isteğin 5, 2. isteğin 4, 3. isteğin 3 saniye sonra cevap dönmesi beklenir.
*/

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

const (
	BASE_URL           = "http://localhost:8000"
	ISTANBUL_REQUESTS  = 11
	ANKARA_REQUESTS    = 3
	ANKARA_INTERVAL    = 1 * time.Second
)

type WeatherResponse struct {
	Location    string  `json:"location"`
	Temperature float64 `json:"temperature"`
}

type TestResult struct {
	RequestNumber int     `json:"request_number"`
	City          string  `json:"city"`
	StartTime     string  `json:"start_time"`
	EndTime       string  `json:"end_time"`
	Duration      float64 `json:"duration"`
	Temperature   float64 `json:"temperature"`
	Success       bool    `json:"success"`
	Error         string  `json:"error,omitempty"`
}

var client = &http.Client{
	Timeout: 30 * time.Second,
}

func makeRequest(city string, requestNum int) TestResult {
	startTime := time.Now()
	
	resp, err := client.Get(fmt.Sprintf("%s/weather?q=%s", BASE_URL, city))
	if err != nil {
		endTime := time.Now()
		duration := endTime.Sub(startTime).Seconds()
		return TestResult{
			RequestNumber: requestNum,
			City:          city,
			StartTime:     startTime.Format("15:04:05.000"),
			EndTime:       endTime.Format("15:04:05.000"),
			Duration:      round(duration, 3),
			Success:       false,
			Error:         err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		endTime := time.Now()
		duration := endTime.Sub(startTime).Seconds()
		return TestResult{
			RequestNumber: requestNum,
			City:          city,
			StartTime:     startTime.Format("15:04:05.000"),
			EndTime:       endTime.Format("15:04:05.000"),
			Duration:      round(duration, 3),
			Success:       false,
			Error:         fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}

	var weatherResp WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		endTime := time.Now()
		duration := endTime.Sub(startTime).Seconds()
		return TestResult{
			RequestNumber: requestNum,
			City:          city,
			StartTime:     startTime.Format("15:04:05.000"),
			EndTime:       endTime.Format("15:04:05.000"),
			Duration:      round(duration, 3),
			Success:       false,
			Error:         err.Error(),
		}
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime).Seconds()
	
	return TestResult{
		RequestNumber: requestNum,
		City:          city,
		StartTime:     startTime.Format("15:04:05.000"),
		EndTime:       endTime.Format("15:04:05.000"),
		Duration:      round(duration, 3),
		Temperature:   weatherResp.Temperature,
		Success:       true,
	}
}

func makeRequestWithDelay(city string, requestNum int, delay time.Duration) TestResult {
	time.Sleep(delay)
	return makeRequest(city, requestNum)
}

func testIstanbul() []TestResult {
	fmt.Println("TEST 1: İstanbul")
	fmt.Println("🎯 Beklenen: 10 istek HEMEN, 1 istek 5 saniye sonra")
	fmt.Printf("⏰ Başlangıç zamanı: %s\n\n", time.Now().Format("15:04:05.000"))

	t0 := time.Now()
	results := make([]TestResult, ISTANBUL_REQUESTS)
	var wg sync.WaitGroup

	for i := 1; i <= ISTANBUL_REQUESTS; i++ {
		wg.Add(1)
		go func(reqNum int) {
			defer wg.Done()
			fmt.Printf("🚀 İstanbul İsteği #%d başlatıldı\n", reqNum)
			results[reqNum-1] = makeRequest("Istanbul", reqNum)
		}(i)
	}

	fmt.Println("\n⏳ İstanbul istekleri tamamlanmayı bekliyor...")
	wg.Wait()

	totalDuration := time.Since(t0).Seconds()

	fmt.Println("\nISTANBUL RESULTS:")
	fmt.Printf("Total duration: %.3f seconds\n\n", totalDuration)

	// Sort by request number
	sort.Slice(results, func(i, j int) bool {
		return results[i].RequestNumber < results[j].RequestNumber
	})

	for _, r := range results {
		if r.Success {
			fmt.Printf("✅ Request #%d: %s → %s | Duration: %.3fs | Temperature: %.1f°C\n",
				r.RequestNumber, r.StartTime, r.EndTime, r.Duration, r.Temperature)
		} else {
			fmt.Printf("❌ Request #%d: %s → %s | Duration: %.3fs | Error: %s\n",
				r.RequestNumber, r.StartTime, r.EndTime, r.Duration, r.Error)
		}
	}

	return results
}

func testAnkara() []TestResult {
	fmt.Println("\n" + "================================================================================\n")
	fmt.Println("TEST 2: Ankara ")
	fmt.Printf("Start time: %s\n\n", time.Now().Format("15:04:05.000"))

	t0 := time.Now()
	results := make([]TestResult, ANKARA_REQUESTS)
	var wg sync.WaitGroup

	for i := 1; i <= ANKARA_REQUESTS; i++ {
		wg.Add(1)
		go func(reqNum int) {
			defer wg.Done()
			delay := time.Duration(reqNum-1) * ANKARA_INTERVAL // 0s, 1s, 2s
			results[reqNum-1] = makeRequestWithDelay("Ankara", reqNum, delay)
		}(i)
	}

	wg.Wait()

	totalDuration := time.Since(t0).Seconds()

	fmt.Println("\nANKARA RESULTS:")
	fmt.Printf("Total duration: %.3f seconds\n\n", totalDuration)

	// Sort by request number
	sort.Slice(results, func(i, j int) bool {
		return results[i].RequestNumber < results[j].RequestNumber
	})

	for _, r := range results {
		if r.Success {
			fmt.Printf("✅ Request #%d: %s → %s | Duration: %.3fs | Temperature: %.1f°C\n",
				r.RequestNumber, r.StartTime, r.EndTime, r.Duration, r.Temperature)
		} else {
			fmt.Printf("❌ Request #%d: %s → %s | Duration: %.3fs | Error: %s\n",
				r.RequestNumber, r.StartTime, r.EndTime, r.Duration, r.Error)
		}
	}

	return results
}

func round(val float64, precision int) float64 {
	ratio := 1.0
	for i := 0; i < precision; i++ {
		ratio *= 10
	}
	return float64(int(val*ratio+0.5)) / ratio
}

func main() {
	fmt.Printf("Base URL: %s\n\n", BASE_URL)
	
	testIstanbul()
	testAnkara()
	
	fmt.Println("✅ Test completed successfully")
}
