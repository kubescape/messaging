package tracer

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kubescape/messaging/pulsar/common/utils"

	"github.com/apache/pulsar-client-go/pulsar"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func BuildAndInjectProducerSpan(producer pulsar.Producer, message *pulsar.ProducerMessage) oteltrace.Span {
	spanName := fmt.Sprintf("PublishTo:%s", producer.Topic())
	tracer := otel.GetTracerProvider().Tracer(
		TracerName,
	)
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), NewProducerMessagePropertiesAdapter(message))
	opts := []oteltrace.SpanStartOption{
		oteltrace.WithAttributes(semconv.MessagingSystemKey.String("pulsar")),
		oteltrace.WithAttributes(semconv.MessagingDestinationNameKey.String(producer.Topic())),
		oteltrace.WithAttributes(semconv.MessagingDestinationKindTopic),
		oteltrace.WithAttributes(semconv.MessagingOperationPublish),
		oteltrace.WithAttributes(semconv.MessagingMessagePayloadSizeBytesKey.Int64(int64(len(message.Payload)))),
		oteltrace.WithAttributes(attribute.KeyValue{Key: "payload", Value: attribute.StringValue(string(message.Payload))}),
		oteltrace.WithSpanKind(oteltrace.SpanKindProducer),
	}
	spanCtx, span := tracer.Start(ctx, spanName, opts...)
	otel.GetTextMapPropagator().Inject(spanCtx, NewProducerMessagePropertiesAdapter(message))
	return span
}

func BuildConsumerSpan(consumer pulsar.Consumer, message pulsar.Message) (context.Context, oteltrace.Span) {
	spanName := fmt.Sprintf("ConsumeFrom%s", message.Topic())
	tracer := otel.GetTracerProvider().Tracer(
		TracerName,
	)
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), NewMessagePropertiesExtractAdapter(message))
	opts := []oteltrace.SpanStartOption{
		oteltrace.WithAttributes(semconv.MessagingSystemKey.String("pulsar")),
		oteltrace.WithAttributes(semconv.MessagingDestinationNameKey.String(message.Topic())),
		oteltrace.WithAttributes(semconv.MessagingDestinationKindTopic),
		oteltrace.WithAttributes(semconv.MessagingOperationReceive),
		oteltrace.WithAttributes(semconv.MessagingConsumerID(consumer.Name())),
		oteltrace.WithAttributes(semconv.MessagingMessageIDKey.String(string(message.ID().Serialize()))),
		oteltrace.WithAttributes(attribute.KeyValue{Key: "subscription", Value: attribute.StringValue(consumer.Subscription())}),
		oteltrace.WithAttributes(semconv.MessagingMessagePayloadSizeBytesKey.Int64(int64(len(message.Payload())))),
		oteltrace.WithAttributes(attribute.KeyValue{Key: "payload", Value: attribute.StringValue(string(message.Payload()))}),
		oteltrace.WithLinks(oteltrace.LinkFromContext(ctx, semconv.OpentracingRefTypeFollowsFrom)),
		oteltrace.WithSpanKind(oteltrace.SpanKindConsumer),
	}
	return tracer.Start(ctx, spanName, opts...)
}

func BuildWorkerSpan(parentSpanCtx, workerCtx context.Context) oteltrace.Span {
	tracer := otel.GetTracerProvider().Tracer(
		TracerName,
	)
	ctxValues := utils.GetContextValues(workerCtx)
	workerName := ctxValues.GetWorkerName()
	spanName := fmt.Sprintf("Worker-%s", workerName)
	opts := []oteltrace.SpanStartOption{
		oteltrace.WithAttributes(attribute.KeyValue{Key: "workerName", Value: attribute.StringValue(workerName)}),
		oteltrace.WithAttributes(semconv.MessagingOperationProcess),
		oteltrace.WithSpanKind(oteltrace.SpanKindInternal),
	}

	if consumer := ctxValues.GetConsumer(); consumer != nil {
		opts = append(opts, oteltrace.WithAttributes(attribute.KeyValue{Key: "consuming:", Value: attribute.StringValue(consumer.Subscription())}))
		spanName = fmt.Sprintf("%s:ConsumeFrom:%s", spanName, consumer.Subscription())
	}
	if producer := ctxValues.GetProducer(); producer != nil {
		opts = append(opts, oteltrace.WithAttributes(attribute.KeyValue{Key: "producing:", Value: attribute.StringValue(producer.Topic())}))
		spanName = fmt.Sprintf("%s:PublishTo:%s", spanName, producer.Topic())
	}
	spanCtx, span := tracer.Start(parentSpanCtx, spanName, opts...)
	otel.GetTextMapPropagator().Inject(spanCtx, ctxValues)
	ctxValues.Store(utils.CurrentSpan, span)
	return span
}

func InjectWorkerSpanToProducerMessage(workerCtx context.Context, message *pulsar.ProducerMessage) {
	if values := utils.GetContextValues(workerCtx); values != nil {
		ctx := otel.GetTextMapPropagator().Extract(context.Background(), values)
		otel.GetTextMapPropagator().Inject(ctx, NewProducerMessagePropertiesAdapter(message))
	}
}

func InjectHTTPRequestSpanToProducerMessage(request *http.Request, message *pulsar.ProducerMessage) {
	otel.GetTextMapPropagator().Inject(request.Context(), NewProducerMessagePropertiesAdapter(message))
}

type MessagePropertiesExtractAdapter struct {
	msg pulsar.Message
}

func NewMessagePropertiesExtractAdapter(message pulsar.Message) *MessagePropertiesExtractAdapter {
	return &MessagePropertiesExtractAdapter{msg: message}
}

// implement TextMapCarrier
func (c *MessagePropertiesExtractAdapter) Get(key string) string {
	return c.msg.Properties()[key]
}

func (c *MessagePropertiesExtractAdapter) Set(key string, value string) {
	panic("extract adapter should not be used for inject")
}

func (c *MessagePropertiesExtractAdapter) Keys() []string {
	keys := make([]string, 0, len(c.msg.Properties()))
	for k := range c.msg.Properties() {
		keys = append(keys, k)
	}
	return keys
}

type ProducerMessagePropertiesAdapter struct {
	msg *pulsar.ProducerMessage
}

func NewProducerMessagePropertiesAdapter(message *pulsar.ProducerMessage) *ProducerMessagePropertiesAdapter {
	if message.Properties == nil {
		message.Properties = make(map[string]string)
	}
	return &ProducerMessagePropertiesAdapter{msg: message}
}

// implement TextMapCarrier
func (c *ProducerMessagePropertiesAdapter) Get(key string) string {
	return c.msg.Properties[key]
}

func (c *ProducerMessagePropertiesAdapter) Set(key string, value string) {
	c.msg.Properties[key] = value
}

func (c *ProducerMessagePropertiesAdapter) Keys() []string {
	keys := make([]string, 0, len(c.msg.Properties))
	for k := range c.msg.Properties {
		keys = append(keys, k)
	}
	return keys
}
