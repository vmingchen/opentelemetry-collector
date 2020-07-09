// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package loggingexporter

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/consumer/pdatautil"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/internal/data"
)

type logDataBuffer struct {
	str strings.Builder
}

func (b *logDataBuffer) logEntry(format string, a ...interface{}) {
	b.str.WriteString(fmt.Sprintf(format, a...))
	b.str.WriteString("\n")
}

func (b *logDataBuffer) logAttr(label string, value string) {
	b.logEntry("    %-15s: %s", label, value)
}

func (b *logDataBuffer) logAttributeMap(label string, am pdata.AttributeMap) {
	if am.Len() == 0 {
		return
	}

	b.logEntry("%s:", label)
	am.ForEach(func(k string, v pdata.AttributeValue) {
		b.logEntry("     -> %s: %s(%s)", k, v.Type().String(), attributeValueToString(v))
	})
}

func (b *logDataBuffer) logStringMap(description string, sm pdata.StringMap) {
	if sm.Len() == 0 {
		return
	}

	b.logEntry("%s:", description)
	sm.ForEach(func(k string, v pdata.StringValue) {
		b.logEntry("     -> %s: %s", k, v.Value())
	})
}

func (b *logDataBuffer) logInstrumentationLibrary(il pdata.InstrumentationLibrary) {
	b.logEntry(
		"InstrumentationLibrary %s %s",
		il.Name(),
		il.Version())
}

func (b *logDataBuffer) logMetricDescriptor(md pdata.MetricDescriptor) {
	if md.IsNil() {
		return
	}

	b.logEntry("Descriptor:")
	b.logEntry("     -> Name: %s", md.Name())
	b.logEntry("     -> Description: %s", md.Description())
	b.logEntry("     -> Unit: %s", md.Unit())
	b.logEntry("     -> Type: %s", md.Type().String())
}

func (b *logDataBuffer) logMetricDataPoints(m pdata.Metric) {
	md := m.MetricDescriptor()
	if md.IsNil() {
		return
	}

	switch md.Type() {
	case pdata.MetricTypeInvalid:
		return
	case pdata.MetricTypeInt64:
		b.logInt64DataPoints(m.Int64DataPoints())
	case pdata.MetricTypeDouble:
		b.logDoubleDataPoints(m.DoubleDataPoints())
	case pdata.MetricTypeMonotonicInt64:
		b.logInt64DataPoints(m.Int64DataPoints())
	case pdata.MetricTypeMonotonicDouble:
		b.logDoubleDataPoints(m.DoubleDataPoints())
	case pdata.MetricTypeHistogram:
		b.logHistogramDataPoints(m.HistogramDataPoints())
	case pdata.MetricTypeSummary:
		b.logSummaryDataPoints(m.SummaryDataPoints())
	}
}

func (b *logDataBuffer) logInt64DataPoints(ps pdata.Int64DataPointSlice) {
	for i := 0; i < ps.Len(); i++ {
		p := ps.At(i)
		if p.IsNil() {
			continue
		}

		b.logEntry("Int64DataPoints #%d", i)
		b.logDataPointLabels(p.LabelsMap())

		b.logEntry("StartTime: %d", p.StartTime())
		b.logEntry("Timestamp: %d", p.Timestamp())
		b.logEntry("Value: %d", p.Value())
	}
}

func (b *logDataBuffer) logDoubleDataPoints(ps pdata.DoubleDataPointSlice) {
	for i := 0; i < ps.Len(); i++ {
		p := ps.At(i)
		if p.IsNil() {
			continue
		}

		b.logEntry("DoubleDataPoints #%d", i)
		b.logDataPointLabels(p.LabelsMap())

		b.logEntry("StartTime: %d", p.StartTime())
		b.logEntry("Timestamp: %d", p.Timestamp())
		b.logEntry("Value: %f", p.Value())
	}
}

func (b *logDataBuffer) logHistogramDataPoints(ps pdata.HistogramDataPointSlice) {
	for i := 0; i < ps.Len(); i++ {
		p := ps.At(i)
		if p.IsNil() {
			continue
		}

		b.logEntry("HistogramDataPoints #%d", i)
		b.logDataPointLabels(p.LabelsMap())

		b.logEntry("StartTime: %d", p.StartTime())
		b.logEntry("Timestamp: %d", p.Timestamp())
		b.logEntry("Count: %d", p.Count())
		b.logEntry("Sum: %f", p.Sum())

		buckets := p.Buckets()
		if buckets.Len() != 0 {
			for i := 0; i < buckets.Len(); i++ {
				bucket := buckets.At(i)
				if bucket.IsNil() {
					continue
				}

				b.logEntry("Buckets #%d, Count: %d", i, bucket.Count())
			}
		}

		bounds := p.ExplicitBounds()
		if len(bounds) != 0 {
			for i, bound := range bounds {
				b.logEntry("ExplicitBounds #%d: %f", i, bound)
			}
		}
	}
}

func (b *logDataBuffer) logSummaryDataPoints(ps pdata.SummaryDataPointSlice) {
	for i := 0; i < ps.Len(); i++ {
		p := ps.At(i)
		if p.IsNil() {
			continue
		}

		b.logEntry("SummaryDataPoints #%d", i)
		b.logDataPointLabels(p.LabelsMap())

		b.logEntry("StartTime: %d", p.StartTime())
		b.logEntry("Timestamp: %d", p.Timestamp())
		b.logEntry("Count: %d", p.Count())
		b.logEntry("Sum: %f", p.Sum())

		percentiles := p.ValueAtPercentiles()
		if percentiles.Len() != 0 {
			for i := 0; i < percentiles.Len(); i++ {
				percentile := percentiles.At(i)
				if percentile.IsNil() {
					continue
				}

				b.logEntry("ValueAtPercentiles #%d, Value: %f, Percentile: %f",
					i, percentile.Value(), percentile.Percentile())
			}
		}
	}
}

func (b *logDataBuffer) logDataPointLabels(labels pdata.StringMap) {
	b.logStringMap("Data point labels", labels)
}

func (b *logDataBuffer) logLogRecord(lr pdata.LogRecord) {
	b.logEntry("Timestamp: %d", lr.Timestamp())
	b.logEntry("Severity: %s", lr.SeverityText())
	b.logEntry("ShortName: %s", lr.ShortName())
	b.logEntry("Body: %s", lr.Body())
	b.logAttributeMap("Attributes", lr.Attributes())
}

func attributeValueToString(av pdata.AttributeValue) string {
	switch av.Type() {
	case pdata.AttributeValueSTRING:
		return av.StringVal()
	case pdata.AttributeValueBOOL:
		return strconv.FormatBool(av.BoolVal())
	case pdata.AttributeValueDOUBLE:
		return strconv.FormatFloat(av.DoubleVal(), 'f', -1, 64)
	case pdata.AttributeValueINT:
		return strconv.FormatInt(av.IntVal(), 10)
	default:
		return fmt.Sprintf("<Unknown OpenTelemetry attribute value type %q>", av.Type())
	}
}

type loggingExporter struct {
	logger *zap.Logger
	debug  bool
}

func (s *loggingExporter) pushTraceData(
	_ context.Context,
	td pdata.Traces,
) (int, error) {

	s.logger.Info("TraceExporter", zap.Int("#spans", td.SpanCount()))

	if !s.debug {
		return 0, nil
	}

	buf := logDataBuffer{}
	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		buf.logEntry("ResourceSpans #%d", i)
		rs := rss.At(i)
		if rs.IsNil() {
			buf.logEntry("* Nil ResourceSpans")
			continue
		}
		if !rs.Resource().IsNil() {
			buf.logAttributeMap("Resource labels", rs.Resource().Attributes())
		}
		ilss := rs.InstrumentationLibrarySpans()
		for j := 0; j < ilss.Len(); j++ {
			buf.logEntry("InstrumentationLibrarySpans #%d", j)
			ils := ilss.At(j)
			if ils.IsNil() {
				buf.logEntry("* Nil InstrumentationLibrarySpans")
				continue
			}
			if !ils.InstrumentationLibrary().IsNil() {
				buf.logInstrumentationLibrary(ils.InstrumentationLibrary())
			}

			spans := ils.Spans()
			for k := 0; k < spans.Len(); k++ {
				buf.logEntry("Span #%d", k)
				span := spans.At(k)
				if span.IsNil() {
					buf.logEntry("* Nil Span")
					continue
				}

				buf.logAttr("Trace ID", span.TraceID().String())
				buf.logAttr("Parent ID", span.ParentSpanID().String())
				buf.logAttr("ID", span.SpanID().String())
				buf.logAttr("Name", span.Name())
				buf.logAttr("Kind", span.Kind().String())
				buf.logAttr("Start time", span.StartTime().String())
				buf.logAttr("End time", span.EndTime().String())
				if !span.Status().IsNil() {
					buf.logAttr("Status code", span.Status().Code().String())
					buf.logAttr("Status message", span.Status().Message())
				}

				buf.logAttributeMap("Attributes", span.Attributes())

				// TODO: Add logging for the rest of the span properties: events, links.
			}
		}
	}
	s.logger.Debug(buf.str.String())

	return 0, nil
}

func (s *loggingExporter) pushMetricsData(
	_ context.Context,
	md pdata.Metrics,
) (int, error) {
	imd := pdatautil.MetricsToInternalMetrics(md)
	s.logger.Info("MetricsExporter", zap.Int("#metrics", imd.MetricCount()))

	if !s.debug {
		return 0, nil
	}

	buf := logDataBuffer{}
	rms := imd.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		buf.logEntry("ResourceMetrics #%d", i)
		rm := rms.At(i)
		if rm.IsNil() {
			buf.logEntry("* Nil ResourceMetrics")
			continue
		}
		if !rm.Resource().IsNil() {
			buf.logAttributeMap("Resource labels", rm.Resource().Attributes())
		}
		ilms := rm.InstrumentationLibraryMetrics()
		for j := 0; j < ilms.Len(); j++ {
			buf.logEntry("InstrumentationLibraryMetrics #%d", j)
			ilm := ilms.At(j)
			if ilm.IsNil() {
				buf.logEntry("* Nil InstrumentationLibraryMetrics")
				continue
			}
			if !ilm.InstrumentationLibrary().IsNil() {
				buf.logInstrumentationLibrary(ilm.InstrumentationLibrary())
			}
			metrics := ilm.Metrics()
			for k := 0; k < metrics.Len(); k++ {
				buf.logEntry("Metric #%d", k)
				metric := metrics.At(k)
				if metric.IsNil() {
					buf.logEntry("* Nil Metric")
					continue
				}

				buf.logMetricDescriptor(metric.MetricDescriptor())
				buf.logMetricDataPoints(metric)
			}
		}
	}

	s.logger.Debug(buf.str.String())

	return 0, nil
}

// NewTraceExporter creates an exporter.TraceExporter that just drops the
// received data and logs debugging messages.
func NewTraceExporter(config configmodels.Exporter, level string, logger *zap.Logger) (component.TraceExporter, error) {
	s := &loggingExporter{
		debug:  level == "debug",
		logger: logger,
	}

	return exporterhelper.NewTraceExporter(
		config,
		s.pushTraceData,
		exporterhelper.WithShutdown(loggerSync(logger)),
	)
}

// NewMetricsExporter creates an exporter.MetricsExporter that just drops the
// received data and logs debugging messages.
func NewMetricsExporter(config configmodels.Exporter, level string, logger *zap.Logger) (component.MetricsExporter, error) {
	s := &loggingExporter{
		debug:  level == "debug",
		logger: logger,
	}

	return exporterhelper.NewMetricsExporter(
		config,
		s.pushMetricsData,
		exporterhelper.WithShutdown(loggerSync(logger)),
	)
}

func loggerSync(logger *zap.Logger) func(context.Context) error {
	return func(context.Context) error {
		// Currently Sync() on stdout and stderr return errors on Linux and macOS,
		// respectively:
		//
		// - sync /dev/stdout: invalid argument
		// - sync /dev/stdout: inappropriate ioctl for device
		//
		// Since these are not actionable ignore them.
		err := logger.Sync()
		if osErr, ok := err.(*os.PathError); ok {
			wrappedErr := osErr.Unwrap()
			switch wrappedErr {
			case syscall.EINVAL, syscall.ENOTSUP, syscall.ENOTTY:
				err = nil
			}
		}
		return err
	}
}

// NewLogExporter creates an exporter.LogExporter that just drops the
// received data and logs debugging messages.
func NewLogExporter(config configmodels.Exporter, level string, logger *zap.Logger) (component.LogExporter, error) {
	s := &loggingExporter{
		debug:  level == "debug",
		logger: logger,
	}

	return exporterhelper.NewLogsExporter(
		config,
		s.pushLogData,
		exporterhelper.WithShutdown(loggerSync(logger)),
	)
}

func (s *loggingExporter) pushLogData(
	_ context.Context,
	ld data.Logs,
) (int, error) {
	s.logger.Info("LogExporter", zap.Int("#logs", ld.LogRecordCount()))

	if !s.debug {
		return 0, nil
	}

	buf := logDataBuffer{}
	rms := ld.ResourceLogs()
	for i := 0; i < rms.Len(); i++ {
		buf.logEntry("ResourceLog #%d", i)
		rm := rms.At(i)
		if rm.IsNil() {
			buf.logEntry("* Nil ResourceLog")
			continue
		}
		if !rm.Resource().IsNil() {
			buf.logAttributeMap("Resource labels", rm.Resource().Attributes())
		}
		lrs := rm.Logs()
		for j := 0; j < lrs.Len(); j++ {
			buf.logEntry("LogRecord #%d", j)
			lr := lrs.At(j)
			if lr.IsNil() {
				buf.logEntry("* Nil LogRecord")
				continue
			}
			buf.logLogRecord(lr)
		}
	}

	s.logger.Debug(buf.str.String())

	return 0, nil
}
