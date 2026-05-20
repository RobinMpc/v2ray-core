package trafficmonitor

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsSnapshot holds per-user traffic rate at a point in time.
type MetricsSnapshot struct {
	Timestamp time.Time `json:"ts"`
	Email     string    `json:"email"`
	UpBPS     float64   `json:"up_bps"`
	DownBPS   float64   `json:"down_bps"`
}

// MetricsSink is the output interface for traffic metrics.
type MetricsSink interface {
	Write(snapshots []MetricsSnapshot) error
	Close() error
}

type userCounter struct {
	uplink   atomic.Int64
	downlink atomic.Int64
}

// TrafficMonitor aggregates per-user traffic and periodically emits BPS snapshots.
type TrafficMonitor struct {
	sink       MetricsSink
	counters   sync.Map // email → *userCounter
	lastValues sync.Map // email → lastSnapshot
	interval   time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
}

type lastSnapshot struct {
	uplink   int64
	downlink int64
}

var globalMonitor *TrafficMonitor
var initOnce sync.Once

// NewMonitor creates a new TrafficMonitor. Pass a nil sink to disable output.
func NewMonitor(sink MetricsSink, interval time.Duration) *TrafficMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	m := &TrafficMonitor{
		sink:     sink,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
	}
	return m
}

// Start begins the background collection goroutine.
func (m *TrafficMonitor) Start() {
	go m.loop()
}

// Close stops the background goroutine and closes the sink.
func (m *TrafficMonitor) Close() error {
	m.cancel()
	if m.sink != nil {
		return m.sink.Close()
	}
	return nil
}

func (m *TrafficMonitor) getOrCreateCounter(email string) *userCounter {
	if v, ok := m.counters.Load(email); ok {
		return v.(*userCounter)
	}
	c := &userCounter{}
	actual, _ := m.counters.LoadOrStore(email, c)
	return actual.(*userCounter)
}

// RecordUplink adds n bytes to the uplink counter for the given email.
func (m *TrafficMonitor) RecordUplink(email string, n int64) {
	m.getOrCreateCounter(email).uplink.Add(n)
}

// RecordDownlink adds n bytes to the downlink counter for the given email.
func (m *TrafficMonitor) RecordDownlink(email string, n int64) {
	m.getOrCreateCounter(email).downlink.Add(n)
}

func (m *TrafficMonitor) loop() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.collectAndWrite()
		}
	}
}

func (m *TrafficMonitor) collectAndWrite() {
	if m.sink == nil {
		return
	}

	now := time.Now()
	intervalSec := m.interval.Seconds()
	var snapshots []MetricsSnapshot

	m.counters.Range(func(key, value interface{}) bool {
		email := key.(string)
		c := value.(*userCounter)
		curUp := c.uplink.Load()
		curDown := c.downlink.Load()

		var prevUp, prevDown int64
		if v, ok := m.lastValues.Load(email); ok {
			ls := v.(*lastSnapshot)
			prevUp = ls.uplink
			prevDown = ls.downlink
		}

		upBPS := float64(curUp-prevUp) / intervalSec
		downBPS := float64(curDown-prevDown) / intervalSec

		m.lastValues.Store(email, &lastSnapshot{uplink: curUp, downlink: curDown})

		snapshots = append(snapshots, MetricsSnapshot{
			Timestamp: now,
			Email:     email,
			UpBPS:     upBPS,
			DownBPS:   downBPS,
		})
		return true
	})

	if len(snapshots) > 0 {
		_ = m.sink.Write(snapshots)
	}
}

// InitMonitor initializes the global singleton. Must be called before GetMonitor.
func InitMonitor(sink MetricsSink, interval time.Duration) {
	initOnce.Do(func() {
		globalMonitor = NewMonitor(sink, interval)
		globalMonitor.Start()
	})
}

// GetMonitor returns the global singleton, or nil if not initialized.
func GetMonitor() *TrafficMonitor {
	return globalMonitor
}
