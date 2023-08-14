package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
 	"go.opentelemetry.io/otel/trace"

 )

var tracer trace.Tracer

type PageSpeedResult struct {
	LighthouseResult struct {
		Categories struct {
			Performance struct {
				Score float64 `json:"score"`
			} `json:"performance"`
		} `json:"categories"`
		Audits struct {
			FirstContentfulPaint struct {
				Score float64 `json:"score"`
			} `json:"first-contentful-paint"`
			SpeedIndex struct {
				Score float64 `json:"score"`
			} `json:"speed-index"`
			LargestContentfulPaint struct {
				DisplayValue string `json:"displayValue"`
			} `json:"largest-contentful-paint"`
			TotalBlockingTime struct {
				Score float64 `json:"score"`
			} `json:"total-blocking-time"`
			CumulativeLayoutShift struct {
				Score float64 `json:"score"`
			} `json:"cumulative-layout-shift"`
		} `json:"audits"`
	} `json:"lighthouseResult"`
}

func getPerformance() (*PageSpeedResult, error) {
	url := "https://www.googleapis.com/pagespeedonline/v5/runPagespeed?url=https://www.marca.com&strategy=MOBILE&category=PERFORMANCE"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PageSpeedResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// httpHandler is an HTTP handler function that is going to be instrumented.
func httpHandler(w http.ResponseWriter, r *http.Request) {
	result, err := getPerformance()
	if err != nil {
		http.Error(w, "Error en la solicitud", http.StatusInternalServerError)
		return
	}
	fmt.Printf("First Contentful Paint: %.2f\n", result.LighthouseResult.Audits.FirstContentfulPaint.Score)
	fmt.Printf("Speed Index: %.2f\n", result.LighthouseResult.Audits.SpeedIndex.Score)
	fmt.Printf("Largest Contentful Paint: %s\n", result.LighthouseResult.Audits.LargestContentfulPaint.DisplayValue)
	fmt.Printf("Total Blocking Time: %.2f\n", result.LighthouseResult.Audits.TotalBlockingTime.Score)
	fmt.Printf("Cumulative Layout Shift: %.2f\n", result.LighthouseResult.Audits.CumulativeLayoutShift.Score)
	// ctx := r.Context()
	ctx, span := tracer.Start(r.Context(), "performance-span")
	sleepy(ctx)
	defer span.End()

}

func main() {
	// Wrap your httpHandler function.

	handler := http.HandlerFunc(httpHandler)
	wrappedHandler := otelhttp.NewHandler(handler, "performance")
	http.Handle("/performance", wrappedHandler)

	log.Fatal(http.ListenAndServe("0.0.0.0:7007", nil))
}

// sleepy mocks work that your application does.
func sleepy(ctx context.Context) {
	_, span := tracer.Start(ctx, "sleep")
	defer span.End()

	sleepTime := 1 * time.Second
	time.Sleep(sleepTime)
	span.SetAttributes(attribute.Int("sleep.duration", int(sleepTime)))
}
