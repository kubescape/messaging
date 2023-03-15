package tracer

import (
	"context"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsar"
)

func NewProducerInterceptors(ctx context.Context) []pulsar.ProducerInterceptor {
	return []pulsar.ProducerInterceptor{
		&TraceProducerInterceptor{ctx: ctx},
	}
}

func NewConsumerInterceptors(ctx context.Context) []pulsar.ConsumerInterceptor {
	return []pulsar.ConsumerInterceptor{
		&TraceConsumerInterceptor{ctx: ctx},
	}
}

type TraceProducerInterceptor struct {
	ctx context.Context
}

func (t *TraceProducerInterceptor) BeforeSend(producer pulsar.Producer, message *pulsar.ProducerMessage) {
	LogTrace(t.ctx, fmt.Sprintf("BeforeSend producer:%s topic:%s message:%s", producer.Name(), producer.Topic(), string(message.Payload)))
	BuildAndInjectProducerSpan(producer, message).End()
}

func (t *TraceProducerInterceptor) OnSendAcknowledgement(producer pulsar.Producer, message *pulsar.ProducerMessage, msgID pulsar.MessageID) {
	LogTrace(t.ctx, fmt.Sprintf("OnSendAcknowledgement producer:%s topic:%s message:%s", producer.Name(), producer.Topic(), string(message.Payload)))
}

type TraceConsumerInterceptor struct {
	ctx context.Context
}

func (t *TraceConsumerInterceptor) BeforeConsume(message pulsar.ConsumerMessage) {
	LogTrace(t.ctx, fmt.Sprintf("BeforeConsume consumer:%s message:%s", message.Consumer.Name(), string(message.Message.Payload())))
}

func (t *TraceConsumerInterceptor) OnAcknowledge(consumer pulsar.Consumer, msgID pulsar.MessageID) {
	LogTrace(t.ctx, fmt.Sprintf("OnAcknowledge consumer:%s message:%s", consumer.Name(), string(msgID.Serialize())))
}

func (t *TraceConsumerInterceptor) OnNegativeAcksSend(consumer pulsar.Consumer, msgIDs []pulsar.MessageID) {
	LogNTraceError(t.ctx, fmt.Sprintf("OnNegativeAcksSend consumer:%s messages:=%v", consumer.Name(), msgIDs), fmt.Errorf("negative Acks Send by consumer:%s", consumer.Name()))
}
