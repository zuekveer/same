package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type CacheMetrics struct {
	Hits   prometheus.Counter
	Misses prometheus.Counter
}

func NewCacheMetrics() *CacheMetrics {
	hits := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total number of cache hits",
	})

	misses := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total number of cache misses",
	})

	return &CacheMetrics{
		Hits:   hits,
		Misses: misses,
	}
}

func (m *Metrics) RegisterCacheMetrics(c *CacheMetrics) {
	m.registry.MustRegister(c.Hits, c.Misses)
}
