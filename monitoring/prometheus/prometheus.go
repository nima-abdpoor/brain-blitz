package prometheus

import "github.com/prometheus/client_golang/prometheus"

var counters []*prometheus.Counter

func GetNewCounter(name, help string) *prometheus.Counter {
	counter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		})
	counters = append(counters, &counter)
	return &counter
}

func IncreaseCounter(counter *prometheus.Counter) {
	(*counter).Inc()
}

func RegisterCounter() {
	for _, counter := range counters {
		prometheus.MustRegister(*counter)
	}
}
