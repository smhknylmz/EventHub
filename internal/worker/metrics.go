package worker

import (
	"context"

	redisadapter "github.com/smhknylmz/EventHub/internal/redis"
	pkgmetrics "github.com/smhknylmz/EventHub/pkg/metrics"
	"go.opentelemetry.io/otel/metric"
)

var (
	DeliveredCounter metric.Int64Counter
	FailedCounter    metric.Int64Counter
	LatencyHistogram metric.Float64Histogram
)

func InitMetrics(queue *redisadapter.Queue) {
	DeliveredCounter, _ = pkgmetrics.Meter.Int64Counter("notifications.delivered")
	FailedCounter, _ = pkgmetrics.Meter.Int64Counter("notifications.failed")
	LatencyHistogram, _ = pkgmetrics.Meter.Float64Histogram("notifications.processing.duration_ms")

	gauge, _ := pkgmetrics.Meter.Int64ObservableGauge("notifications.queue.depth")
	pkgmetrics.Meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		depth, err := queue.TotalDepth(ctx)
		if err != nil {
			return err
		}
		o.ObserveInt64(gauge, depth)
		return nil
	}, gauge)
}
