package showq

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Time defines a timestamp encoded as epoch seconds in JSON.
type Timestamp time.Time

// MarshalJSON is used to convert the timestamp to JSON.
func (t *Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(*t).Unix(), 10)), nil
}

// UnmarshalJSON is used to convert the timestamp from JSON.
func (t *Timestamp) UnmarshalJSON(s []byte) (err error) {
	r := string(s)
	q, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = time.Unix(q, 0)
	return nil
}

// String returns timestamp string
func (t Timestamp) String() string {
	return strconv.FormatInt(time.Time(t).Unix(), 10)
}

// Recipient and the reason for the delay.
type Recipient struct {
	Address     string  `json:"address"`
	DelayReason *string `json:"delay_reason"`
}

// Messages in the sendmail-style queue.
// See: http://www.postfix.org/postqueue.1.html
type Message struct {
	QueueName    string      `json:"queue_name"`
	QueueId      string      `json:"queue_id"`
	ArrivalTime  Timestamp   `json:"arrival_time"`
	MessageSize  uint64      `json:"message_size"`
	ForcedExpire bool        `json:"forced_expire"`
	Sender       string      `json:"sender"`
	Recipients   []Recipient `json:"recipients"`
}

// Bytes returns a message converted into bytes
func (m *Message) Bytes() []byte {
	var arr []string
	arr = append(arr, fmt.Sprintf("queue_name\000%s", m.QueueName))
	arr = append(arr, fmt.Sprintf("queue_id\000%s", m.QueueId))
	arr = append(arr, fmt.Sprintf("time\000%d", time.Time(m.ArrivalTime).Unix()))
	arr = append(arr, fmt.Sprintf("size\000%d", m.MessageSize))
	arr = append(arr, fmt.Sprintf("forced_expire\000%t", m.ForcedExpire))
	arr = append(arr, fmt.Sprintf("sender\000%s", m.Sender))
	for _, recipient := range m.Recipients {
		arr = append(arr, fmt.Sprintf("recipient\000%s", recipient.Address))
		if recipient.DelayReason != nil {
			arr = append(arr, fmt.Sprintf("reason\000%s", *recipient.DelayReason))
		}
	}
	return []byte(strings.Join(arr, "\000") + "\000\000")
}

// A parsing error occurs when an unexpected string of characters is encountered.
type ParseError struct {
	message string
	line    string
}

// Line returns a string of lines that failed to be parsed.
func (e *ParseError) Line() string {
	return e.line
}

// Error returns an error string.
func (e *ParseError) Error() string {
	return e.message
}

// A Reader reads message from a showq.
type Reader struct {
	r  *bufio.Reader
	mu sync.Mutex
}

// NewReader returns a new Reader that reads from r.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		r:  bufio.NewReader(r),
		mu: sync.Mutex{},
	}
}

// readLine retrieves a single message string from showq.
func (r *Reader) readLine() ([]byte, error) {
	// format: key1\0value1\0key2\0value2\0\0key1\0value1\0\0\0
	var buf []byte
	for {
		line, err := r.r.ReadSlice(0)
		if err == bufio.ErrBufferFull {
			var b []byte
			for err == bufio.ErrBufferFull {
				line, err = r.r.ReadSlice(0)
				b = append(b, line...)
			}
			line = b
		}
		if err == io.EOF {
			return nil, io.EOF
		}
		if len(buf) == 0 && len(line) == 1 && line[0] == 0 {
			return nil, io.EOF
		}
		if len(line) == 1 && line[0] == 0 {
			break
		}

		buf = append(buf, line...)
	}

	return bytes.TrimSuffix(buf, []byte{0}), nil
}

// Read reads one record (a slice of fields) from r.
func (r *Reader) Read() (*Message, error) {
	r.mu.Lock()
	line, err := r.readLine()
	r.mu.Unlock()
	if err != nil {
		return nil, err
	}
	record := bytes.Split(line, []byte{0})
	if (len(record) % 2) != 0 {
		return nil, &ParseError{
			message: "An unexpected error occurred in ShowQ's parsing",
			line:    string(line),
		}
	}

	message := &Message{}
	for i := 0; i < len(record); i += 2 {
		key := string(record[i])
		value := string(record[i+1])
		switch key {
		case "queue_name":
			message.QueueName = value
		case "queue_id":
			message.QueueId = value
		case "time":
			ts, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, err
			}
			t := time.Unix(ts, 0)
			message.ArrivalTime = Timestamp(t)
		case "size":
			size, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, err
			}
			message.MessageSize = size
		case "sender":
			message.Sender = value
		case "forced_expire":
			b, err := strconv.ParseBool(value)
			if err != nil {
				return nil, err
			}
			message.ForcedExpire = b
		case "recipient":
			message.Recipients = append(message.Recipients, Recipient{Address: value})

		case "reason":
			recipient := &message.Recipients[len(message.Recipients)-1]
			recipient.DelayReason = &value

		default:
			return nil, &ParseError{
				message: fmt.Sprintf("There are keys that do not match `queue_name`, `queue_id`, `time`, `size`, `sender`, `forced_expire`, `recipient`, `reason`, actuality they are `%s`.", key),
				line:    string(line),
			}
		}
	}
	return message, nil
}
