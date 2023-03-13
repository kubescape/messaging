package pulsarconnector

import (
	"context"
	"fmt"

	"github.com/kubescape/pulsar-connector/common/tracer"
	"github.com/kubescape/pulsar-connector/common/utils"

	"github.com/apache/pulsar-client-go/pulsar"
	"go.opentelemetry.io/otel/codes"
)

type PubSubWorkersManager struct {
	ctx               context.Context
	consumer          pulsar.Consumer
	ready2receiveChan chan bool
	receiveMsgs       chan pulsar.Message
	producer          pulsar.Producer
}

func NewPubSubWorkersManager(ctx context.Context, numOfWorkers int, consumer pulsar.Consumer, producer pulsar.Producer) *PubSubWorkersManager {
	p := &PubSubWorkersManager{
		ctx:               ctx,
		consumer:          consumer,
		ready2receiveChan: make(chan bool, numOfWorkers),
		receiveMsgs:       make(chan pulsar.Message, numOfWorkers),
		producer:          producer,
	}
	go p.start()
	return p
}

func (p *PubSubWorkersManager) start() {
	if p.producer != nil {
		defer p.producer.Close()
	}
	defer p.consumer.Close()
	for {
		select {
		case <-p.ctx.Done():
			tracer.LogTrace(p.ctx, "pub-sub workers manager done")
			return
		case <-p.ready2receiveChan:
			if msg, err := p.consumer.Receive(p.ctx); err == nil {
				p.receiveMsgs <- msg
				continue
			} else if p.ctx.Err() == nil {
				tracer.LogNTraceError(p.ctx, "failed to receive message", err)
			}
		}
	}
}

func (p *PubSubWorkersManager) Receive(workerCtx context.Context) (pulsar.Message, error) {
	p.ready2receiveChan <- true
	select {
	case <-workerCtx.Done():
		return nil, workerCtx.Err()
	case <-p.ctx.Done():
		return nil, p.ctx.Err()
	case msg := <-p.receiveMsgs:
		spanCtx, span := tracer.BuildConsumerSpan(p.consumer, msg)
		span.End()
		tracer.BuildWorkerSpan(spanCtx, workerCtx)
		return msg, nil
	}
}

func (p *PubSubWorkersManager) Ack(workerCtx context.Context, msg pulsar.Message) error {
	err := p.consumer.Ack(msg)
	if values := utils.GetContextValues(workerCtx); values != nil {
		if span := values.GetAndDeleteCurrentSpan(); span != nil {
			defer span.End()
			if err != nil {
				tracer.LogNTraceError(workerCtx, "failed to ack message from topic "+msg.Topic(), err)
				span.SetStatus(codes.Error, err.Error())
			} else {
				log := "ack message from topic " + msg.Topic()
				tracer.LogTrace(workerCtx, log)
				span.SetStatus(codes.Ok, log)
			}
		}
	}
	return err
}

func (p *PubSubWorkersManager) Nack(workerCtx context.Context, msg pulsar.Message) {
	p.consumer.Nack(msg)
	if values := utils.GetContextValues(workerCtx); values != nil {
		if span := values.GetAndDeleteCurrentSpan(); span != nil {
			defer span.End()
			span.SetStatus(codes.Error, "message nack")
		}
	}
}

func (p *PubSubWorkersManager) Send(workerCtx context.Context, msg *pulsar.ProducerMessage) error {
	if p.producer == nil {
		return fmt.Errorf("producer is not set")
	}
	tracer.InjectWorkerSpanToProducerMessage(workerCtx, msg)
	_, err := p.producer.Send(p.ctx, msg)
	return err
}

func (p *PubSubWorkersManager) Done() bool {
	return p.ctx.Err() != nil
}
