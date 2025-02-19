package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// RepositoryMetrics เป็น interface สำหรับ metrics ของ repositories ทั้งหมด
type RepositoryMetrics interface {
	// เพิ่ม methods ตามที่ต้องการ
	Register()
}

// HTTPMetrics สำหรับเก็บ metrics ของ HTTP requests
type HTTPMetrics struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
}

// NewHTTPMetrics สร้าง metrics สำหรับ HTTP
func NewHTTPMetrics() *HTTPMetrics {
	m := &HTTPMetrics{
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
	}
	return m
}

// DatabaseMetrics สำหรับเก็บ metrics ของ database
type DatabaseMetrics struct {
	ConnectionsOpen *prometheus.GaugeVec
	QueryDuration   *prometheus.HistogramVec
	QueriesTotal    *prometheus.CounterVec
}

func NewDatabaseMetrics(name string) *DatabaseMetrics {
	m := &DatabaseMetrics{
		ConnectionsOpen: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: name + "_db_connections_open",
				Help: "Number of open database connections",
			},
			[]string{"database"},
		),
		QueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    name + "_db_query_duration_seconds",
				Help:    "Duration of database queries",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		),
		QueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: name + "_db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),
	}
	return m
}

// CacheMetrics สำหรับเก็บ metrics ของ cache
type CacheMetrics struct {
	HitsTotal         prometheus.Counter
	MissesTotal       prometheus.Counter
	OperationDuration *prometheus.HistogramVec
}

func NewCacheMetrics() *CacheMetrics {
	m := &CacheMetrics{
		HitsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total number of cache hits",
			},
		),
		MissesTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total number of cache misses",
			},
		),
		OperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "cache_operation_duration_seconds",
				Help:    "Duration of cache operations",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
	}
	return m
}

// MessageMetrics สำหรับเก็บ metrics ของ message broker
type MessageMetrics struct {
	MessagesPublished *prometheus.CounterVec
	PublishDuration   *prometheus.HistogramVec
}

func NewMessageMetrics() *MessageMetrics {
	m := &MessageMetrics{
		MessagesPublished: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "messages_published_total",
				Help: "Total number of messages published",
			},
			[]string{"topic", "status"},
		),
		PublishDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "message_publish_duration_seconds",
				Help:    "Duration of message publish operations",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"topic"},
		),
	}
	return m
}

// SearchMetrics สำหรับเก็บ metrics ของ search operations
type SearchMetrics struct {
	SearchesTotal   *prometheus.CounterVec
	SearchDuration  *prometheus.HistogramVec
	ResultsReturned *prometheus.HistogramVec
}

var (
	searchMetricsSingleton    *SearchMetrics
	searchMetricsSingletonMux sync.Mutex
)

// NewSearchMetrics creates a new SearchMetrics instance or returns the existing one
func NewSearchMetrics() *SearchMetrics {
	searchMetricsSingletonMux.Lock()
	defer searchMetricsSingletonMux.Unlock()

	if searchMetricsSingleton != nil {
		return searchMetricsSingleton
	}

	searchMetricsSingleton = &SearchMetrics{
		SearchDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "search_duration_seconds",
				Help: "Duration of search operations in seconds",
			},
			[]string{"index"},
		),
		SearchesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "searches_total",
				Help: "Total number of searches performed",
			},
			[]string{"index", "status"},
		),
		ResultsReturned: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "search_results_returned",
				Help: "Number of results returned by search operations",
			},
			[]string{"index"},
		),
	}

	return searchMetricsSingleton
}
