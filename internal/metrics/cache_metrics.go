package metrics

import "github.com/prometheus/client_golang/prometheus"

type CacheMetrics struct {
	Hits      prometheus.Counter
	Misses    prometheus.Counter
	Evictions prometheus.Counter
}

func NewCacheMetrics(reg prometheus.Registerer) *CacheMetrics {
	hits := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total number of cache hits",
	})
	misses := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total number of cache misses",
	})
	evictions := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_evictions_total",
		Help: "Total number of evicted entries",
	})

	reg.MustRegister(hits, misses, evictions)

	return &CacheMetrics{
		Hits:      hits,
		Misses:    misses,
		Evictions: evictions,
	}
}
