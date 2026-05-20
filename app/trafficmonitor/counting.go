package trafficmonitor

import (
	"github.com/v2fly/v2ray-core/v5/common/buf"
)

// CountingReader wraps a buf.Reader and records bytes read toward traffic counters.
type CountingReader struct {
	reader  buf.Reader
	monitor *TrafficMonitor
	email   string
	isUplink bool
}

// NewCountingReader creates a CountingReader.
// Set isUplink=true to count uplink traffic; false for downlink.
func NewCountingReader(reader buf.Reader, monitor *TrafficMonitor, email string, isUplink bool) *CountingReader {
	return &CountingReader{
		reader:   reader,
		monitor:  monitor,
		email:    email,
		isUplink: isUplink,
	}
}

func (r *CountingReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	mb, err := r.reader.ReadMultiBuffer()
	if mb != nil && !mb.IsEmpty() {
		n := int64(mb.Len())
		if r.isUplink {
			r.monitor.RecordUplink(r.email, n)
		} else {
			r.monitor.RecordDownlink(r.email, n)
		}
	}
	return mb, err
}

// CountingWriter wraps a buf.Writer and records bytes written toward traffic counters.
type CountingWriter struct {
	writer   buf.Writer
	monitor  *TrafficMonitor
	email    string
	isUplink bool
}

// NewCountingWriter creates a CountingWriter.
// Set isUplink=true to count uplink traffic; false for downlink.
func NewCountingWriter(writer buf.Writer, monitor *TrafficMonitor, email string, isUplink bool) *CountingWriter {
	return &CountingWriter{
		writer:   writer,
		monitor:  monitor,
		email:    email,
		isUplink: isUplink,
	}
}

func (w *CountingWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	if mb != nil && !mb.IsEmpty() {
		n := int64(mb.Len())
		if w.isUplink {
			w.monitor.RecordUplink(w.email, n)
		} else {
			w.monitor.RecordDownlink(w.email, n)
		}
	}
	return w.writer.WriteMultiBuffer(mb)
}
