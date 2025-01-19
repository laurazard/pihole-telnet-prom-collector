package collector

import (
	"time"

	"github.com/laurazard/pihole-telnet-prom-collector/pihole"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type PiHoleCollector struct {
	queryBucketsDesc *prometheus.Desc
	queryGaugeDesc   *prometheus.Desc

	queryBuckets *prometheus.HistogramVec

	piClient      pihole.Client
	lastCollected time.Time
}

func NewPiHoleCollector(url string) *PiHoleCollector {
	piClient := pihole.TelnetClient{
		URL: url,
	}

	histOpts := prometheus.HistogramVecOpts{
		HistogramOpts: prometheus.HistogramOpts{
			Namespace:                    "pihole",
			Name:                         "query_duration_ms_buckets",
			Buckets:                      []float64{1, 3, 5, 10, 20, 50, 100, 200, 1000},
			NativeHistogramBucketFactor:  1.2,
			NativeHistogramZeroThreshold: 0.2,
		},
		VariableLabels: bucketLabels,
	}
	hist := prometheus.V2.NewHistogramVec(histOpts)

	return &PiHoleCollector{
		piClient:      &piClient,
		lastCollected: time.Now(),
		queryBucketsDesc: prometheus.NewDesc(
			"pihole_query_duration_ms_buckets",
			"Histogram of DNS query duration.",
			[]string{"type", "reply", "status", "upstream"}, nil,
		),
		queryGaugeDesc: prometheus.NewDesc(
			"pihole_query_duration_ms",
			"DNS query duration, represented as a gauge.",
			[]string{"type", "reply", "status", "upstream", "client", "domain"}, nil),
		queryBuckets: hist,
	}
}

func (c *PiHoleCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.queryBucketsDesc
	ch <- c.queryGaugeDesc
}

func (c *PiHoleCollector) Collect(ch chan<- prometheus.Metric) {
	gaugeOpts := prometheus.GaugeVecOpts{
		GaugeOpts: prometheus.GaugeOpts{
			Namespace: "pihole",
			Name:      "query_duration_ms",
		},
		VariableLabels: gaugeLabels,
	}
	metres := prometheus.V2.NewGaugeVec(gaugeOpts)

	collectionTime := time.Now()
	queries, err := c.piClient.GetQueries(c.lastCollected)
	if err != nil {
		logrus.WithError(err).Errorf("failed to collect")
		return
	}

	for _, q := range queries {
		c.queryBuckets.With(queryBucketLabels(q)).Observe(q.DelayMs)
		metres.With(queryGaugeLabels(q)).Set(q.DelayMs)
	}

	c.queryBuckets.Collect(ch)
	metres.Collect(ch)
	c.lastCollected = collectionTime
}

var (
	bucketLabels = prometheus.ConstrainedLabels{
		{Name: "type"},
		{Name: "status"},
		{Name: "reply"},
		{Name: "upstream"},
	}

	gaugeLabels = prometheus.ConstrainedLabels{
		{Name: "type"},
		{Name: "client"},
		{Name: "status"},
		{Name: "reply"},
		{Name: "upstream"},
		{Name: "domain"},
	}
)

func queryBucketLabels(q pihole.Query) prometheus.Labels {
	return prometheus.Labels{
		"type":     q.QueryType,
		"status":   q.Status.String(),
		"reply":    q.ReplyType.String(),
		"upstream": q.Upstream,
	}
}

func queryGaugeLabels(q pihole.Query) prometheus.Labels {
	return prometheus.Labels{
		"type":     q.QueryType,
		"status":   q.Status.String(),
		"reply":    q.ReplyType.String(),
		"client":   q.ClientIP,
		"upstream": q.Upstream,
		"domain":   q.Domain,
	}
}
