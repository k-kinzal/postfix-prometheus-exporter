package mock

import (
	"github.com/k-kinzal/postfix-prometheus-exporter/postfix/encoding/showq"
	"time"
)

type ShowqMessageGenFunc func() []showq.Message

func ShowqMessageGen(num int) ShowqMessageGenFunc {
	return func() []showq.Message {
		messages := make([]showq.Message, num)
		for i := 0; i < num; i++ {
			messages[i] =
				showq.Message{
					QueueName:    "deferred",
					QueueId:      "09229268B721",
					ArrivalTime:  showq.Timestamp(time.Unix(0, 0)),
					MessageSize:  0,
					ForcedExpire: false,
					Sender:       "foo@example.com",
					Recipients: []showq.Recipient{
						{
							Address:     "bar@example.jp",
							DelayReason: nil,
						},
					},
				}
		}
		return messages
	}
}