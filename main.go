// main package
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/imdhruva/stock-picker/docs"
	"github.com/rs/cors"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// TimeSeriesDaily holds the metadata and a list of timeseries data points for a stock
type TimeSeriesDaily struct {
	MetaData   MetaData                  `json:"Meta Data"`
	TimeSeries map[string]StockDataPoint `json:"Time Series (Daily)"`
}

// MetaData holds symbol/ticker for a specific stock
type MetaData struct {
	Symbol string `json:"2. Symbol"`
}

// StockDataPoint holds the closing value for a stock
type StockDataPoint struct {
	Close string `json:"4. close"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// StockData represents the stock data response
type StockData struct {
	Days    map[string]float64 `json:"days"`
	Average float64            `json:"average"`
}

// @title Stock Service API
// @description API to retrieve closing prices of a specific stock
// @version 1.0
// @host localhost:8080
// @BasePath /
func main() {
	r := gin.Default()

	// init swagger
	docs.SwaggerInfo.Title = "Stock Picker"
	docs.SwaggerInfo.Description = "API that fetches average of closing price of a given stock ticker for the last given nDays."
	docs.SwaggerInfo.Version = "1.0"

	// Enable CORS
	corsMiddleware := cors.Default()
	r.Use(func(c *gin.Context) {
		corsMiddleware.HandlerFunc(c.Writer, c.Request)
		c.Next()
	})

	r.GET("/stock", getStockData)

	// Swagger documentation
	url := ginSwagger.URL("/swagger/doc.json")
	//  Swagger JSON file path
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s\n", port)
	err := r.Run(":" + port)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// @Summary Get stock data
// @Description Get the last NDAYS days of data along with the average closing price
// @Accept json
// @Produce json
// @Param symbol query string true "Stock symbol"
// @Param nDays query int true "Number of days"
// @Success 200 {object} StockData
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /stock [get]
func getStockData(c *gin.Context) {
	url := os.Getenv("ALPHA_VANTAGE_URL")
	if url == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "ALPHA_VANTAGE_URL environment variable is not set"})
		return
	}

	symbol := os.Getenv("SYMBOL")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "SYMBOL environment variable is not set"})
		return
	}

	nDaysStr := os.Getenv("NDAYS")
	if nDaysStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "NDAYS environment variable is not set"})
		return
	}

	nDaysInt, err := strconv.Atoi(nDaysStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Failed to convert NDAYS to in. Err: %s", err)})
		return
	}

	nDays, err := time.ParseDuration(fmt.Sprintf("-%dh", nDaysInt*24))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Failed to parse NDAYS. Err: %s", err)})
		return
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "API_KEY environment variable not set"})
	}

	quoteURL := fmt.Sprintf("%s/query?apikey=%s&function=TIME_SERIES_DAILY&symbol=%s", url, apiKey, symbol)

	resp, err := http.Get(quoteURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve stock quote"})
		return
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to close the request body.Err: %s", err)})
			return
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Wrong status code received. Expected: %v, Got: %v", http.StatusOK, resp.StatusCode)})
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to read stock quote response"})
		return
	}

	var data TimeSeriesDaily
	if err := json.Unmarshal(body, &data); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to parse stock quote JSON. Err: %s", err)})
		return
	}

	days, err := parseNDays(nDays, data.TimeSeries)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to parse NDays. Err: %s", err)})
		return
	}

	average := calculateAverageClosingPrice(days)

	response := StockData{
		Days:    days,
		Average: average,
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, response)
}

func parseNDays(nDays time.Duration, timeSeries map[string]StockDataPoint) (map[string]float64, error) {
	days := make(map[string]float64)

	now := time.Now()
	for date, stockData := range timeSeries {
		stockDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			continue
		}

		if !stockDate.Before(now.Add(nDays)) {
			closeF64, err := strconv.ParseFloat(stockData.Close, 64)
			if err != nil {
				return days, fmt.Errorf("invalid value for 'StockDataPoint.Time Series (Daily).4. close'. Expected: string, got %T", stockData.Close)
			}
			days[date] = closeF64
		}

	}

	return days, nil
}

func calculateAverageClosingPrice(days map[string]float64) float64 {
	total := 0.0
	count := 0

	for _, close := range days {
		total += close
		count++
	}

	if count > 0 {
		return total / float64(count)
	}

	return 0.0
}
