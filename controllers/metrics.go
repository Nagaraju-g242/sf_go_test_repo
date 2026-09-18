package controllers

import "github.com/prometheus/client_golang/prometheus"

var (
	ReconcileTotal  = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "snowflake_operator_reconcile_total", Help: "Total reconcile attempts."}, []string{"kind"})
	ReconcileErrors = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "snowflake_operator_reconcile_errors_total", Help: "Total reconcile errors."}, []string{"kind"})
	DriftRepairs    = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "snowflake_operator_drift_repairs_total", Help: "Resources recreated after being found missing in Snowflake."}, []string{"kind"})
	Duration        = prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "snowflake_operator_reconcile_duration_seconds", Help: "Reconcile duration in seconds."}, []string{"kind"})
	Managed         = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "snowflake_operator_managed_resources", Help: "Number of managed Kubernetes resources currently observed."}, []string{"kind"})
)

func init() {
	prometheus.MustRegister(ReconcileTotal, ReconcileErrors, DriftRepairs, Duration, Managed)
}
