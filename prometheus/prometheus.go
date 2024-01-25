package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

var (
	myCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "my_counter",
		Help: "This is my counter",
	})

	myGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "my_gauge",
		Help: "This is my gauge",
	})

	myHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "my_histogram",
		Help:    "This is my histogram",
		Buckets: prometheus.LinearBuckets(20, 5, 5), // Start at 20, 5 wide, 5 buckets
	})

	mySummary = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "my_summary",
		Help:       "This is my summary",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	})
)

func init() {
	// Register metrics with Prometheus's default registry
	prometheus.MustRegister(myCounter)
	prometheus.MustRegister(myGauge)
	prometheus.MustRegister(myHistogram)
	prometheus.MustRegister(mySummary)
}

func main() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		myCounter.Add(float64(i))
	}
}
