package app

import "github.com/prometheus/client_golang/prometheus"

// AppMetrics holds app-level metrics. We will connect them to real data later.
type AppMetrics struct {
	RequestsTotal  prometheus.Counter
	LoginSuccess   prometheus.Counter
	LoginFailure   prometheus.Counter
	ActiveSessions prometheus.Gauge
}

func NewAppMetrics() *AppMetrics {
	return &AppMetrics{
		RequestsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "lab_app_requests_total",
			Help: "Total HTTP requests handled by the application.",
		}),
		LoginSuccess: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "lab_app_login_success_total",
			Help: "Total successful logins.",
		}),
		LoginFailure: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "lab_app_login_fail_total",
			Help: "Total failed logins.",
		}),
		ActiveSessions: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "lab_app_active_sessions",
			Help: "Number of active user sessions.",
		}),
	}
}

func (m *AppMetrics) Register() {
	prometheus.MustRegister(
		m.RequestsTotal,
		m.LoginSuccess,
		m.LoginFailure,
		m.ActiveSessions,
	)
}
