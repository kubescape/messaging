package utils

import (
	"context"
	"sync"

	"github.com/apache/pulsar-client-go/pulsar"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type ContextValues struct {
	*sync.Map
}
type ContextKey string

const (
	Values      ContextKey = "notification-values"
	Producer    ContextKey = "producer"
	Consumer    ContextKey = "consumer"
	Logger      ContextKey = "logger"
	WorkerName  ContextKey = "worker-name"
	CurrentSpan ContextKey = "current-span"
)

func NewContextWithValues(ctx context.Context, workerName string) context.Context {
	ctxValues := &ContextValues{Map: &sync.Map{}}
	parentValues := GetContextValues(ctx)
	if parentValues != nil {
		parentValues.Range(func(key, value interface{}) bool {
			if key != WorkerName {
				ctxValues.Store(key, value)
			}
			return true
		})
	}
	ctxValues.Store(WorkerName, workerName)
	ctxValues.Store(Logger, ctxValues.GetLogger().With(zap.String("worker-name", workerName)))
	return context.WithValue(ctx, Values, ctxValues)
}

func GetContextValues(ctx context.Context) *ContextValues {
	if val := ctx.Value(Values); val != nil {
		return val.(*ContextValues)
	}
	return nil
}

func GetContextLogger(ctx context.Context) *zap.Logger {
	if values := GetContextValues(ctx); values != nil {
		return values.GetLogger()
	}
	return zap.L()
}

func SetContextProducer(ctx context.Context, publisher pulsar.Producer) {
	if values := GetContextValues(ctx); values != nil {
		values.Store(Producer, publisher)
		values.buildProducerLogger(publisher)
	}
}

func GetContextSpan(ctx context.Context) oteltrace.Span {
	if values := GetContextValues(ctx); values != nil {
		return values.GetCurrentSpan()
	}
	return nil
}

func SetContextConsumer(ctx context.Context, consumer pulsar.Consumer) {
	if values := GetContextValues(ctx); values != nil {
		values.Store(Consumer, consumer)
		values.buildConsumerLogger(consumer)
	}
}

func (c *ContextValues) GetLogger() *zap.Logger {
	if z, ok := c.Load(Logger); z != nil && ok {
		return z.(*zap.Logger)
	}
	return zap.L()
}

func (c *ContextValues) GetCurrentSpan() oteltrace.Span {
	if s, ok := c.Load(CurrentSpan); s != nil && ok {
		span := s.(oteltrace.Span)
		if span.SpanContext().IsValid() {
			return span
		}
	}
	return nil
}

func (c *ContextValues) GetAndDeleteCurrentSpan() oteltrace.Span {
	if s, ok := c.LoadAndDelete(CurrentSpan); s != nil && ok {
		span := s.(oteltrace.Span)
		if span.SpanContext().IsValid() {
			return span
		}
	}
	return nil
}

func (c *ContextValues) GetProducer() pulsar.Producer {
	if p, ok := c.Load(Producer); p != nil && ok {
		return p.(pulsar.Producer)
	}
	return nil
}

func (c *ContextValues) GetConsumer() pulsar.Consumer {
	if c, ok := c.Load(Consumer); c != nil && ok {
		return c.(pulsar.Consumer)
	}
	return nil
}

func (c *ContextValues) GetWorkerName() string {
	if i, ok := c.Load(WorkerName); i != nil && ok {
		return i.(string)
	}
	return ""
}

func (c *ContextValues) buildProducerLogger(publisher pulsar.Producer) {
	fields := []zap.Field{}
	if topic := publisher.Topic(); topic != "" {
		fields = append(fields, zap.String("topic", topic))
	}
	if name := publisher.Name(); name != "" {
		fields = append(fields, zap.String("publisher-name", name))
	}
	c.Store(Logger, c.GetLogger().With(fields...))
}

func (c *ContextValues) buildConsumerLogger(Consumer pulsar.Consumer) {
	fields := []zap.Field{}
	if topic := Consumer.Subscription(); topic != "" {
		fields = append(fields, zap.String("subscription", topic))
	}
	if name := Consumer.Name(); name != "" {
		fields = append(fields, zap.String("consumer-name", name))
	}
	c.Store(Logger, c.GetLogger().With(fields...))
}

func (c *ContextValues) Get(key string) string {
	if v, ok := c.Load(key); v != nil && ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c *ContextValues) Set(key string, value string) {
	c.Store(key, value)
}

func (c *ContextValues) Keys() []string {
	keys := []string{}
	c.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys
}
