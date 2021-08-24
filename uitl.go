package go_prom_pprof_utils

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"
)

var (
	ns          string      // namespace for prometheus
	su          *statsUtils // main struct
	onceForHttp sync.Once   // for http serve
)

type statsUtils struct {
	ctx        context.Context
	metrics    sync.Map
	registerer prometheus.Registerer
	server     *http.Server
	logger     *logrus.Logger
}

func NewStatUtils(ctx context.Context, namespace string, usePprof bool, addr net.Addr, logger *logrus.Logger) error {
	if su != nil {
		return errors.New("create stats struct")
	}

	if ctx == nil || namespace == "" || addr == nil || logger == nil {
		return errors.New("invalid argument")
	}

	ns = namespace

	registerer := prometheus.NewRegistry()
	handler := promhttp.InstrumentMetricHandler(registerer, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}))

	router := mux.NewRouter()

	if usePprof {
		router.PathPrefix("/debug/").Handler(http.DefaultServeMux)
		router.HandleFunc("/debug/pprof/", pprof.Index)
		router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		router.HandleFunc("/debug/pprof/profile", pprof.Profile)
		router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		router.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	router.Handle("/metrics", handler)

	su = &statsUtils{
		ctx:        ctx,
		metrics:    sync.Map{},
		registerer: registerer,
		server: &http.Server{
			Addr:    addr.String(),
			Handler: router,
		},
		logger: logger,
	}

	return nil
}

func Serve() {
	onceForHttp.Do(func() {
		gorillaClosable := make(chan struct{})
		go func() {
			select {
			case <-su.ctx.Done():
				if err := su.server.Shutdown(context.Background()); err != nil {
					su.logger.WithError(err).Error("gorilla server shutdown err")
				}
				close(gorillaClosable)
			}
		}()

		if err := su.server.ListenAndServe(); err != nil {
			su.logger.WithError(err).Error("ListenAndServe")
		}

		<-gorillaClosable

		su.logger.Info("gorilla websocket server terminate")
	})
}

func Increase(key string, types ...string) {
	if value, ok := su.metrics.Load(key); ok {
		switch value.(type) {
		case prometheus.Gauge:
			value.(prometheus.Gauge).Inc()
		case *prometheus.GaugeVec:
			if types != nil {
				value.(*prometheus.GaugeVec).WithLabelValues(types...).Inc()
			}
		case prometheus.Counter:
			value.(prometheus.Counter).Inc()
		case *prometheus.CounterVec:
			if types != nil {
				value.(*prometheus.CounterVec).WithLabelValues(types...).Inc()
			}
		}
	}
}

func Decrease(key string, types ...string) {
	if value, ok := su.metrics.Load(key); ok {
		switch value.(type) {
		case prometheus.Gauge:
			value.(prometheus.Gauge).Dec()
		case *prometheus.GaugeVec:
			if types != nil {
				value.(*prometheus.GaugeVec).WithLabelValues(types...).Dec()
			}
		}
	}
}

func IncreaseHistogram(key string, startTime time.Time, types ...string) {
	if value, ok := su.metrics.Load(key); ok {
		switch value.(type) {
		case prometheus.Histogram:
			value.(prometheus.Histogram).Observe(time.Since(startTime).Seconds())
		case *prometheus.HistogramVec:
			if types != nil {
				value.(*prometheus.HistogramVec).WithLabelValues(types...).Observe(time.Since(startTime).Seconds())
			}
		}
	}
}

func AddGauge(name string, types []string) {
	if types == nil {
		newGauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      name,
		})
		prometheus.MustRegister(newGauge)
		su.metrics.Store(name, newGauge)
		su.registerer.MustRegister(newGauge)
	} else {
		newGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      name,
		}, types)
		prometheus.MustRegister(newGauge)
		su.metrics.Store(name, newGauge)
		su.registerer.MustRegister(newGauge)
	}
}

func AddCounter(name string, types []string) {
	if types == nil {
		newCounter := prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: ns,
			Name:      name,
		})
		prometheus.MustRegister(newCounter)
		su.metrics.Store(name, newCounter)
		su.registerer.MustRegister(newCounter)
	} else {
		newCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: ns,
			Name:      name,
		}, types)
		prometheus.MustRegister(newCounter)
		su.metrics.Store(name, newCounter)
		su.registerer.MustRegister(newCounter)
	}
}

func AddSummary(name string, types []string) {
	if types == nil {
		newSummary := prometheus.NewSummary(prometheus.SummaryOpts{
			Namespace: ns,
			Name:      name,
		})
		prometheus.MustRegister(newSummary)
		su.metrics.Store(name, newSummary)
		su.registerer.MustRegister(newSummary)
	} else {
		newSummary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace: ns,
			Name:      name,
		}, types)

		prometheus.MustRegister(newSummary)
		su.metrics.Store(name, newSummary)
		su.registerer.MustRegister(newSummary)
	}
}

func AddHistogram(name string, buckets []float64, types []string) {
	if types == nil {
		newHistogram := prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: ns,
			Name:      name,
			Buckets:   buckets,
		})
		prometheus.MustRegister(newHistogram)
		su.metrics.Store(name, newHistogram)
		su.registerer.MustRegister(newHistogram)
	} else {
		newHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: ns,
			Name:      name,
			Buckets:   buckets,
		}, types)

		prometheus.MustRegister(newHistogram)
		su.metrics.Store(name, newHistogram)
		su.registerer.MustRegister(newHistogram)
	}
}
