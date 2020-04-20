package collector

import (
	"encoding/json"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/jasonlvhit/gocron"
	"github.com/k-kinzal/postfix-prometheus-exporter/postfix"
	"github.com/k-kinzal/postfix-prometheus-exporter/postfix/encoding/showq"
	"github.com/k-kinzal/postfix-prometheus-exporter/util"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	golog "log"
	"sync"
	"time"
)

// Simple lock implementation to lock gocron jobs.
type locker struct {
	mutexs map[string]sync.Mutex
}

// Lock is locking job.
func (l *locker) Lock(key string) (bool, error) {
	m, ok := l.mutexs[key]
	if !ok {
		m = sync.Mutex{}
	}
	m.Lock()
	l.mutexs[key] = m
	return true, nil
}

// Lock is unlocking job.
func (l *locker) Unlock(key string) error {
	m, ok := l.mutexs[key]
	if !ok {
		return nil
	}
	m.Unlock()
	return nil
}

// PostfixQueueCollectScheduler to collect statistics for Postfix queue.
type PostfixQueueCollectScheduler struct {
	collector *PostfixQueueCollector
}

// Collector returns the Collector of prometheus.
func (s *PostfixQueueCollectScheduler) Collector() prometheus.Collector {
	return s.collector
}

// Collect collects queue statistics from the postqueue.
func (s *PostfixQueueCollectScheduler) Collect() {
	level.Debug(s.collector.logger).Log("msg", "Start collecting")
	now := time.Now()

	s.collector.mu.Lock()
	defer s.collector.mu.Unlock()

	s.collector.sizeBytesHistogram.Reset()
	s.collector.ageSecondsHistogram.Reset()
	s.collector.scrapeSuccessGauge.Reset()
	s.collector.scrapeDurationGauge.Reset()

	cnt := 0
	mu := sync.Mutex{}
	err := s.collector.postqueue.EachProduce(func(message *showq.Message) {
		for i := 0; i < len(message.Recipients); i++ {
			message.Recipients[i].Address = util.EmailMask(message.Recipients[i].Address)
		}
		b, _ := json.Marshal(message)
		level.Debug(s.collector.logger).Log("msg", "Collected items", "item", b)

		mu.Lock()
		defer mu.Unlock()

		s.collector.sizeBytesHistogram.WithLabelValues(message.QueueName).Observe(float64(message.MessageSize))
		s.collector.ageSecondsHistogram.WithLabelValues(message.QueueName).Observe(now.Sub(time.Time(message.ArrivalTime)).Seconds())
		cnt++
	})

	if err != nil {
		if e, ok := err.(*showq.ParseError); ok {
			level.Error(s.collector.logger).Log("err", err, "line", util.EmailMask(e.Line()))
		} else {
			level.Error(s.collector.logger).Log("err", err)
		}
		s.collector.scrapeSuccessGauge.WithLabelValues("postfix_queue").Set(0)
	} else {
		s.collector.scrapeSuccessGauge.WithLabelValues("postfix_queue").Set(1)
	}
	s.collector.scrapeDurationGauge.WithLabelValues("postfix_queue").Set(time.Now().Sub(now).Seconds())

	_, nextTime := gocron.NextRun()
	level.Debug(s.collector.logger).Log("msg", "Finish collecting", "length", cnt, "duration", time.Now().Sub(now).Seconds(), "next", nextTime)
}

// Start starts to collect statistics of postfix queue.
// Because collection starts after interval_seconds, if you want to collect immediately, please call Collect after start.
func (s *PostfixQueueCollectScheduler) Start(intervalSeconds uint64) chan bool {
	level.Debug(s.collector.logger).Log("msg", "Starting postfix queue collector", "interval", intervalSeconds)
	golog.SetOutput(ioutil.Discard) // disable gocron log
	gocron.SetLocker(&locker{make(map[string]sync.Mutex)})
	gocron.Every(intervalSeconds).Seconds().Lock().Do(s.Collect)
	return gocron.Start()
}

// NewPostfixQueueCollectScheduler returns new PostfixQueueCollectScheduler.
func NewPostfixQueueCollectScheduler(q *postfix.PostQueue, logger log.Logger) *PostfixQueueCollectScheduler {
	return &PostfixQueueCollectScheduler{
		collector: &PostfixQueueCollector{
			postqueue: q,
			logger:    logger,
			mu:        sync.Mutex{},
			sizeBytesHistogram: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Namespace: "postfix",
					Subsystem: "queue",
					Name:      "size_bytes",
					Help:      "Total message size in the queue.",
					Buckets:   []float64{1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9},
				},
				[]string{"queue_name"}),
			ageSecondsHistogram: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Namespace: "postfix",
					Subsystem: "queue",
					Name:      "age_seconds",
					Help:      "Age of messages in the queue, in seconds.",
					Buckets:   []float64{1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8},
				},
				[]string{"queue_name"}),

			scrapeDurationGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "postfix",
					Subsystem: "scope",
					Name:      "collector_duration_seconds",
					Help:      "postfix_exporter: Duration of a collector scrape.",
				},
				[]string{"collector"}),
			scrapeSuccessGauge: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "postfix",
					Subsystem: "scope",
					Name:      "collector_success",
					Help:      "postfix_exporter: Whether a collector succeeded.",
				},
				[]string{"collector"}),
		},
	}
}

// PostfixQueueCollector to collect statistics of postfix queue in Prometheus format
type PostfixQueueCollector struct {
	postqueue *postfix.PostQueue
	logger    log.Logger
	mu        sync.Mutex

	// metrics
	sizeBytesHistogram  *prometheus.HistogramVec
	ageSecondsHistogram *prometheus.HistogramVec
	scrapeDurationGauge *prometheus.GaugeVec
	scrapeSuccessGauge  *prometheus.GaugeVec
}

// Describe implements the prometheus.Collector interface.
func (c *PostfixQueueCollector) Describe(ch chan<- *prometheus.Desc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ageSecondsHistogram.Describe(ch)
	c.sizeBytesHistogram.Describe(ch)
	c.scrapeDurationGauge.Describe(ch)
	c.scrapeSuccessGauge.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (c *PostfixQueueCollector) Collect(ch chan<- prometheus.Metric) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ageSecondsHistogram.Collect(ch)
	c.sizeBytesHistogram.Collect(ch)
	c.scrapeDurationGauge.Collect(ch)
	c.scrapeSuccessGauge.Collect(ch)
}
