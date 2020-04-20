package postfix_test

import (
	"context"
	"fmt"
	"github.com/k-kinzal/postfix-prometheus-exporter/postfix"
	"github.com/k-kinzal/postfix-prometheus-exporter/postfix/encoding/showq"
	"github.com/k-kinzal/postfix-prometheus-exporter/test/mock"
	"testing"
	"time"
)

func ExamplePostQueue() {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	showqPath, _ := mock.Serve(ctx, mock.ShowqMessageGen(1))

	queue := postfix.NewPostQueue(&postfix.PostQueueOpt{ShowqPath: showqPath})
	messages, err := queue.Produce()
	if err != nil {
		panic(err)
	}
	fmt.Println(messages)
	// Output: [{deferred 09229268B721 0 0 false foo@example.com [{bar@example.jp <nil>}]}]
}

func TestPostQueue_Produce(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	showqPath, expected := mock.Serve(ctx, mock.ShowqMessageGen(3))

	queue := postfix.NewPostQueue(&postfix.PostQueueOpt{ShowqPath: showqPath})
	messages, err := queue.Produce()
	if err != nil {
		t.Fatal(err)
	}
	for i, message := range messages {
		if message.QueueName != expected[i].QueueName {
			t.Errorf("expected `%v`, but actual is `%v`", expected[i].QueueName, message.QueueName)
		}
		if message.QueueID != expected[i].QueueID {
			t.Errorf("expected `%v`, but actual is `%v`", expected[i].QueueID, message.QueueID)
		}
		if time.Time(message.ArrivalTime).Unix() != time.Time(message.ArrivalTime).Unix() {
			t.Errorf("expected `%v`, but actual is `%v`", time.Time(message.ArrivalTime).Unix(), time.Time(message.ArrivalTime).Unix())
		}
		if message.MessageSize != expected[i].MessageSize {
			t.Errorf("expected `%v`, but actual is `%v`", expected[i].MessageSize, message.MessageSize)
		}
		if message.ForcedExpire != expected[i].ForcedExpire {
			t.Errorf("expected `%v`, but actual is `%v`", expected[i].ForcedExpire, message.ForcedExpire)
		}
		if message.Recipients[0].Address != expected[i].Recipients[0].Address {
			t.Errorf("expected `%v`, but actual is `%v`", expected[i].Recipients[0].Address, message.Recipients[0].Address)
		}
		if fmt.Sprintf("%v", message.Recipients[0].DelayReason) != fmt.Sprintf("%v", message.Recipients[0].DelayReason) {
			t.Errorf("expected `%v`, but actual is `%v`", expected[i].Recipients[0].DelayReason, message.Recipients[0].DelayReason)
		}
	}
}

func TestPostQueue_EachPoduce(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	showqPath, _ := mock.Serve(ctx, mock.ShowqMessageGen(3))

	cnt := 0
	queue := postfix.NewPostQueue(&postfix.PostQueueOpt{ShowqPath: showqPath})
	err := queue.EachProduce(func(message *showq.Message) {
		cnt++
	})
	if err != nil {
		t.Fatal(err)
	}
	if cnt != 3 {
		t.Errorf("expected `3`, but actual is `%v`", cnt)
	}
}

// BenchmarkPostQueue_Produce-8   	  659408	      2037 ns/op
func BenchmarkPostQueue_Produce(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	showqPath, _ := mock.Serve(ctx, mock.ShowqMessageGen(b.N))

	b.ResetTimer()

	queue := postfix.NewPostQueue(&postfix.PostQueueOpt{ShowqPath: showqPath})
	_, err := queue.Produce()
	if err != nil {
		b.Fatal(err)
	}
}

// BenchmarkPostQueue_EachProduce-8   	  904740	      1180 ns/op
func BenchmarkPostQueue_EachProduce(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	showqPath, _ := mock.Serve(ctx, mock.ShowqMessageGen(b.N))

	b.ResetTimer()

	queue := postfix.NewPostQueue(&postfix.PostQueueOpt{ShowqPath: showqPath})
	err := queue.EachProduce(func(message *showq.Message) {

	})
	if err != nil {
		b.Fatal(err)
	}
}