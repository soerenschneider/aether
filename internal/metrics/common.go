package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace             = "aether"
	subsystemNotification = "notification"
)

var (
	NotificationValidationErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystemNotification,
		Name:      "validation_errors_total",
		Help:      "Total errors validating notifications",
	}, []string{"service"})
)
