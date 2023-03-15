package tracer

import (
	"context"
	"fmt"

	"github.com/kubescape/messaging/pulsar/common/utils"

	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const TracerName = "github.com/kubescape/messaging"

func LogTrace(c context.Context, msg string) {
	var fields []zapcore.Field
	if span := utils.GetContextSpan(c); span != nil {
		span.AddEvent(msg)
		fields = append(fields, zap.String("traceID", span.SpanContext().TraceID().String()))
		fields = append(fields, zap.String("spanID", span.SpanContext().SpanID().String()))
	}
	log(msg, nil, c, fields...)
}

func LogNTraceError(c context.Context, msg string, err error) {
	var fields []zapcore.Field
	if span := utils.GetContextSpan(c); span != nil {
		span.SetStatus(codes.Error, fmt.Sprintf("%s error:%v", msg, err))
		fields = append(fields, zap.String("traceID", span.SpanContext().TraceID().String()))
		fields = append(fields, zap.String("spanID", span.SpanContext().SpanID().String()))
	}
	log(msg, err, c, fields...)
}

func LogNTraceEnterExit(c context.Context, msg string) func() {
	LogTrace(c, msg)
	return func() {
		LogTrace(c, msg+" completed")
	}
}

func log(msg string, err error, c context.Context, fields ...zapcore.Field) {
	logger := utils.GetContextLogger(c)
	if err != nil {
		logger.Error(msg, append(fields, zap.Error(err))...)
	} else {
		logger.Info(msg, fields...)
	}
}
