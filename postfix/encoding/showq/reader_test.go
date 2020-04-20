package showq_test

import (
	"bytes"
	"fmt"
	"github.com/k-kinzal/postfix-prometheus-exporter/postfix/encoding/showq"
	"github.com/k-kinzal/postfix-prometheus-exporter/test/mock"
	"io"
	"testing"
	"time"
)

func ExampleReader() {
	expected := mock.ShowqMessageGen(1)()
	reader := showq.NewReader(bytes.NewReader(append(expected[0].Bytes(), 0)))
	message, err := reader.Read()
	if err != nil {
		panic(err)
	}
	fmt.Println(message)
	// Output: &{deferred 09229268B721 0 0 false foo@example.com [{bar@example.jp <nil>}]}
}

func TestReader_Read(t *testing.T) {
	expected := mock.ShowqMessageGen(1)()[0]
	reader := showq.NewReader(bytes.NewReader(append(expected.Bytes(), 0)))
	message, err := reader.Read()
	if err != nil {
		t.Fatal(err)
	}
	if message.QueueName != expected.QueueName {
		t.Errorf("expected `%v`, but actual is `%v`", expected.QueueName, message.QueueName)
	}
	if message.QueueId != expected.QueueId {
		t.Errorf("expected `%v`, but actual is `%v`", expected.QueueId, message.QueueId)
	}
	if time.Time(message.ArrivalTime).Unix() != time.Time(message.ArrivalTime).Unix() {
		t.Errorf("expected `%v`, but actual is `%v`", time.Time(message.ArrivalTime).Unix(), time.Time(message.ArrivalTime).Unix())
	}
	if message.MessageSize != expected.MessageSize {
		t.Errorf("expected `%v`, but actual is `%v`", expected.MessageSize, message.MessageSize)
	}
	if message.ForcedExpire != expected.ForcedExpire {
		t.Errorf("expected `%v`, but actual is `%v`", expected.ForcedExpire, message.ForcedExpire)
	}
	if message.Recipients[0].Address != expected.Recipients[0].Address {
		t.Errorf("expected `%v`, but actual is `%v`", expected.Recipients[0].Address, message.Recipients[0].Address)
	}
	if fmt.Sprintf("%v", message.Recipients[0].DelayReason) != fmt.Sprintf("%v", message.Recipients[0].DelayReason) {
		t.Errorf("expected `%v`, but actual is `%v`", expected.Recipients[0].DelayReason, message.Recipients[0].DelayReason)
	}
}

func TestReader_ReadMultipleRecipients(t *testing.T) {
	reason := "none"
	expected := mock.ShowqMessageGen(1)()[0]
	expected.Recipients = append(expected.Recipients, showq.Recipient{
		Address:     "bar@example.com",
		DelayReason: &reason,
	})

	reader := showq.NewReader(bytes.NewReader(append(expected.Bytes(), 0)))
	message, err := reader.Read()
	if err != nil {
		t.Fatal(err)
	}
	if message.QueueName != expected.QueueName {
		t.Errorf("expected `%v`, but actual is `%v`", expected.QueueName, message.QueueName)
	}
	if message.QueueId != expected.QueueId {
		t.Errorf("expected `%v`, but actual is `%v`", expected.QueueId, message.QueueId)
	}
	if time.Time(message.ArrivalTime).Unix() != time.Time(message.ArrivalTime).Unix() {
		t.Errorf("expected `%v`, but actual is `%v`", time.Time(message.ArrivalTime).Unix(), time.Time(message.ArrivalTime).Unix())
	}
	if message.MessageSize != expected.MessageSize {
		t.Errorf("expected `%v`, but actual is `%v`", expected.MessageSize, message.MessageSize)
	}
	if message.ForcedExpire != expected.ForcedExpire {
		t.Errorf("expected `%v`, but actual is `%v`", expected.ForcedExpire, message.ForcedExpire)
	}
	if message.Recipients[0].Address != expected.Recipients[0].Address {
		t.Errorf("expected `%v`, but actual is `%v`", expected.Recipients[0].Address, message.Recipients[0].Address)
	}
	if fmt.Sprintf("%v", message.Recipients[0].DelayReason) != fmt.Sprintf("%v", message.Recipients[0].DelayReason) {
		t.Errorf("expected `%v`, but actual is `%v`", expected.Recipients[0].DelayReason, message.Recipients[0].DelayReason)
	}
	if message.Recipients[1].Address != expected.Recipients[1].Address {
		t.Errorf("expected `%v`, but actual is `%v`", expected.Recipients[1].Address, message.Recipients[1].Address)
	}
	if fmt.Sprintf("%v", message.Recipients[1].DelayReason) != fmt.Sprintf("%v", message.Recipients[1].DelayReason) {
		t.Errorf("expected `%v`, but actual is `%v`", expected.Recipients[1].DelayReason, message.Recipients[1].DelayReason)
	}
}

func TestReader_ReadMultipleMessage(t *testing.T) {
	expected := mock.ShowqMessageGen(1)()[0]
	buf := append(expected.Bytes(), expected.Bytes()...)
	reader := showq.NewReader(bytes.NewReader(append(buf, 0)))
	for {
		message, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
		if message.QueueName != expected.QueueName {
			t.Errorf("expected `%v`, but actual is `%v`", expected.QueueName, message.QueueName)
		}
		if message.QueueId != expected.QueueId {
			t.Errorf("expected `%v`, but actual is `%v`", expected.QueueId, message.QueueId)
		}
		if time.Time(message.ArrivalTime).Unix() != time.Time(message.ArrivalTime).Unix() {
			t.Errorf("expected `%v`, but actual is `%v`", time.Time(message.ArrivalTime).Unix(), time.Time(message.ArrivalTime).Unix())
		}
		if message.MessageSize != expected.MessageSize {
			t.Errorf("expected `%v`, but actual is `%v`", expected.MessageSize, message.MessageSize)
		}
		if message.ForcedExpire != expected.ForcedExpire {
			t.Errorf("expected `%v`, but actual is `%v`", expected.ForcedExpire, message.ForcedExpire)
		}
		if message.Recipients[0].Address != expected.Recipients[0].Address {
			t.Errorf("expected `%v`, but actual is `%v`", expected.Recipients[0].Address, message.Recipients[0].Address)
		}
		if fmt.Sprintf("%v", message.Recipients[0].DelayReason) != fmt.Sprintf("%v", message.Recipients[0].DelayReason) {
			t.Errorf("expected `%v`, but actual is `%v`", expected.Recipients[0].DelayReason, message.Recipients[0].DelayReason)
		}
	}
}

// BenchmarkReader_Read-8   	  909477	      1236 ns/op
func BenchmarkReader_Read(b *testing.B) {
	expected := mock.ShowqMessageGen(b.N)()
	var buf []byte
	for i := 0; i < b.N; i++ {
		buf = append(buf, expected[i].Bytes()...)
	}
	reader := showq.NewReader(bytes.NewReader(append(buf, 0)))

	b.ResetTimer()
	for {
		_, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			b.Fatal(err)
		}
	}
}