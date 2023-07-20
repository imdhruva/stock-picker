package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestStockService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stock Service Suite")
}

var _ = Describe("Stock Service", func() {
	var (
		router *gin.Engine
		server *httptest.Server
	)
	const (
		testAPIKey = "o2837weuhfkhasdk"
		testSymbol = "MSFT"
		testNDAYS  = "7"
	)

	Describe("parseNDays", func() {
		It("should parse a valid integer value for nDays", func() {
			// set input value
			nDays := 7
			nDuration := fmt.Sprintf("-%dh", nDays*24)
			duration, err := time.ParseDuration(nDuration)
			Expect(err).ToNot(HaveOccurred())

			today := time.Now()

			timeSeries := map[string]StockDataPoint{
				today.Format("2006-01-02"):                   {Close: "100.0"},
				today.AddDate(0, 0, -1).Format("2006-01-02"): {Close: "200.0"},
				today.AddDate(0, 0, -2).Format("2006-01-02"): {Close: "150.0"},
				today.AddDate(0, 0, -3).Format("2006-01-02"): {Close: "180.0"},
				today.AddDate(0, 0, -4).Format("2006-01-02"): {Close: "250.0"},
				today.AddDate(0, 0, -5).Format("2006-01-02"): {Close: "220.0"},
				today.AddDate(0, 0, -6).Format("2006-01-02"): {Close: "300.0"},
				today.AddDate(0, 0, -7).Format("2006-01-02"): {Close: "320.0"},
				today.AddDate(0, 0, -8).Format("2006-01-02"): {Close: "340.0"},
			}
			result, err := parseNDays(duration, timeSeries)

			expectedTimeSeries := map[string]float64{
				today.Format("2006-01-02"):                   100,
				today.AddDate(0, 0, -1).Format("2006-01-02"): 200.0,
				today.AddDate(0, 0, -2).Format("2006-01-02"): 150.0,
				today.AddDate(0, 0, -3).Format("2006-01-02"): 180.0,
				today.AddDate(0, 0, -4).Format("2006-01-02"): 250.0,
				today.AddDate(0, 0, -5).Format("2006-01-02"): 220.0,
				today.AddDate(0, 0, -6).Format("2006-01-02"): 300.0,
			}

			// Assert the expected result and error
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeEquivalentTo(expectedTimeSeries))
		})
	})

	Describe("calculateAverageClosingPrice", func() {
		It("should calculate average closing price over nDays", func() {
			today := time.Now()
			timeSeries := map[string]float64{
				today.Format("2006-01-02"):                   100,
				today.AddDate(0, 0, -1).Format("2006-01-02"): 200.0,
				today.AddDate(0, 0, -2).Format("2006-01-02"): 150.0,
				today.AddDate(0, 0, -3).Format("2006-01-02"): 180.0,
				today.AddDate(0, 0, -4).Format("2006-01-02"): 250.0,
			}
			result := calculateAverageClosingPrice(timeSeries)
			Expect(result).To(Equal(float64(176)))
		})
	})

	BeforeSuite(func() {
		// Create a mock HTTP server
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set the desired response for each test case
			switch r.URL.Path {
			case "/query":
				// Check the query parameters and respond accordingly
				query := r.URL.Query()
				symbol := query.Get("symbol")
				apiKey := query.Get("apikey")

				today := time.Now()

				// Check the query parameters to determine the response
				if symbol == testSymbol && apiKey == testAPIKey {
					// Create a mock response JSON
					response := TimeSeriesDaily{
						TimeSeries: map[string]StockDataPoint{
							today.Format("2006-01-02"):                   {Close: "100.0"},
							today.AddDate(0, 0, -1).Format("2006-01-02"): {Close: "120.0"},
							today.AddDate(0, 0, -2).Format("2006-01-02"): {Close: "140.0"},
							today.AddDate(0, 0, -3).Format("2006-01-02"): {Close: "160.0"},
							today.AddDate(0, 0, -4).Format("2006-01-02"): {Close: "180.0"},
							today.AddDate(0, 0, -5).Format("2006-01-02"): {Close: "200.0"},
							today.AddDate(0, 0, -6).Format("2006-01-02"): {Close: "220.0"},
						},
					}
					responseJSON, err := json.Marshal(response)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					// Set the Content-Type header to indicate JSON response
					w.Header().Set("Content-Type", "application/json")

					_, err = w.Write([]byte(responseJSON))
					Expect(err).ToNot(HaveOccurred())

				} else {
					// Respond with an error for other cases
					w.WriteHeader(http.StatusBadRequest)
					_, err := w.Write([]byte(`{"error": "Invalid request"}`))
					Expect(err).ToNot(HaveOccurred())
				}
			default:
				w.WriteHeader(http.StatusNotFound)
				_, err := w.Write([]byte(`{"error": "Not found"}`))
				Expect(err).ToNot(HaveOccurred())
			}
		}))
	})

	AfterSuite(func() {
		// Close the mock HTTP server after all tests
		server.Close()
	})

	BeforeEach(func() {
		// Set up the Gin router
		router = gin.Default()
		router.GET("/stock", getStockData)
	})

	It("should return the stock data and average closing price", func() {
		// Set the required environment variables
		err := os.Setenv("SYMBOL", testSymbol)
		Expect(err).ToNot(HaveOccurred())
		err = os.Setenv("NDAYS", testNDAYS)
		Expect(err).ToNot(HaveOccurred())
		err = os.Setenv("API_KEY", testAPIKey)
		Expect(err).ToNot(HaveOccurred())

		// Set the mock Alpha Vantage URL to the test server URL
		err = os.Setenv("ALPHA_VANTAGE_URL", server.URL)
		Expect(err).ToNot(HaveOccurred())

		// Create a mock HTTP request
		request, err := http.NewRequest(http.MethodGet, "/stock", nil)
		Expect(err).ToNot(HaveOccurred())

		// Create a mock HTTP response recorder
		response := httptest.NewRecorder()

		// Dispatch the request to the Gin router
		router.ServeHTTP(response, request)

		// Read the response body
		body := response.Body.Bytes()

		Expect(response.Code).To(Equal(http.StatusOK))

		var jsonResponse StockData
		err = json.Unmarshal(body, &jsonResponse)
		Expect(err).ToNot(HaveOccurred())

		// Assert the expected response values
		expectedNDays := 7
		Expect(len(jsonResponse.Days)).To(Equal(expectedNDays))
		Expect(jsonResponse.Average).ToNot(Equal(0.0))
	})
})
