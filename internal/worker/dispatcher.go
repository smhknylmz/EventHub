package worker

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/smhknylmz/EventHub/internal/notification"
	redisadapter "github.com/smhknylmz/EventHub/internal/redis"
)

var priorityWeights = map[string]int64{
	notification.PriorityHigh:   5,
	notification.PriorityNormal: 3,
	notification.PriorityLow:    1,
}

type Dispatcher struct {
	queue     *redisadapter.Queue
	processor *Processor
	consumer  string
}

func NewDispatcher(queue *redisadapter.Queue, processor *Processor, consumer string) *Dispatcher {
	return &Dispatcher{
		queue:     queue,
		processor: processor,
		consumer:  consumer,
	}
}

func (d *Dispatcher) Start(ctx context.Context) {
	if err := d.queue.CreateConsumerGroups(ctx); err != nil {
		log.WithError(err).Fatal("failed to create consumer groups")
	}

	var wg sync.WaitGroup

	for _, key := range redisadapter.AllStreamKeys() {
		wg.Add(1)
		go func(stream string) {
			defer wg.Done()
			d.consumeStream(ctx, stream)
		}(key)
	}

	wg.Wait()
}

func (d *Dispatcher) consumeStream(ctx context.Context, stream string) {
	priority := redisadapter.PriorityFromStream(stream)
	count := priorityWeights[priority]

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		notifications, err := d.queue.Read(ctx, stream, d.consumer, count)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.WithError(err).WithField("stream", stream).Error("failed to read from stream")
			time.Sleep(time.Second)
			continue
		}

		for _, n := range notifications {
			d.processor.Process(ctx, n)
		}
	}
}
