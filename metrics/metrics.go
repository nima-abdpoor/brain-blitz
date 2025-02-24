package metrics

import "github.com/prometheus/client_golang/prometheus"

var FailedClaimCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "failed_get_claims_count",
		Help: "Error In Getting Claims From Echo Context Count",
	},
)

var FailedMatchedUserCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "failed_match_users_count",
		Help: "Failed to match Users Count",
	},
)

var SucceedMatchedUserCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "success_match_users_count",
		Help: "Successful User match Count",
	},
)

var TestMetric = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "Test_Count",
		Help: "Successful Test Metric Count",
	},
)

func InitMetrics() {
	prometheus.MustRegister(FailedClaimCounter, FailedMatchedUserCounter, SucceedMatchedUserCounter, TestMetric)
}
