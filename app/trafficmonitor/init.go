package trafficmonitor

import (
	"os"
	"strconv"
	"time"
)

func init() {
	sinkType := os.Getenv("V2RAY_TRAFFIC_SINK")
	if sinkType == "" {
		sinkType = "file"
	}

	intervalSec := 5
	if v := os.Getenv("V2RAY_TRAFFIC_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			intervalSec = n
		}
	}

	var sink MetricsSink
	switch sinkType {
	case "file":
		path := os.Getenv("V2RAY_TRAFFIC_LOG_PATH")
		if path == "" {
			path = "/var/log/v2ray/traffic.json"
		}
		var err error
		sink, err = NewFileSink(path)
		if err != nil {
			// File sink is unavailable; monitor will be a no-op.
			return
		}
	default:
		return
	}

	InitMonitor(sink, time.Duration(intervalSec)*time.Second)
}
