package metrics

import "github.com/prometheus/client_golang/prometheus"

// @ 1- Metrics variables declration
var HttpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	
	// defining up counter -> with these options set on it
	Name: "go_app_total_http_requests", // name for the counter in promthues
	Help: "total number of http requests being handled by the Backend", //its desc
	
},
[]string{"path","status"}, //* labels for tracing

)

// gauge -> up-down counter <- for tracking ws clients connections
var ActiveConnections = prometheus.NewGauge(prometheus.GaugeOpts{
	
	// defining -> gauge counter name 
	Name:"go_app_total_active_websocket_connections",
	Help: "current number of active websocket connections ",
})


// @ 2 - Metrics variables registration
// initializing and registering those into the theus (short for promtheus)
func init() {
	prometheus.MustRegister(HttpRequestsTotal)
	prometheus.MustRegister(ActiveConnections)
}

//@ 3 - expose endpoint where all metrics are served, resgister in the routes
//@ 4 - use those registered variables to increment or decrement as when required 