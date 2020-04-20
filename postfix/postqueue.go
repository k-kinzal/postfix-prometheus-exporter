package postfix

import (
	"github.com/k-kinzal/postfix-prometheus-exporter/postfix/encoding/showq"
	"io"
	"net"
	"runtime"
	"sync"
)

// See: http://www.postfix.org/postqueue.1.html
type PostQueueOpt struct {
	// configDir string // FIXME: want to parse main.cf and read the queue_directory.
	ShowqPath string
}

// Postfix user interface for queue management.
// See: http://www.postfix.org/postqueue.1.html
type PostQueue struct {
	opt *PostQueueOpt
}

// connectShowq returns connection to showq.
func (q *PostQueue) connectShowq() (net.Conn, error) {
	path := q.opt.ShowqPath
	if path == "" {
		path = "/var/spool/postfix/public/showq"
	}
	return net.Dial("unix", path)
}

// Each will produce a traditional sendmail-style queue list of messages per unit.
func (q *PostQueue) EachProduce(fn func(message *showq.Message)) error {
	conn, err := q.connectShowq()
	if err != nil {
		return err
	}
	defer conn.Close()

	var er error
	reader := showq.NewReader(conn)
	wg := sync.WaitGroup{}
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for er == nil {
				message, e := reader.Read()
				if e != nil {
					er = e
					return
				}
				fn(message)
			}
		}()
	}
	wg.Wait()

	if er != io.EOF {
		return er
	} else {
		return nil
	}
}

// Produce a traditional sendmail-style queue listing.
func (q *PostQueue) Produce() ([]showq.Message, error) {
	var messages []showq.Message

	conn, err := q.connectShowq()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	reader := showq.NewReader(conn)
	for {
		message, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		messages = append(messages, *message)
	}
	return messages, nil
}

// NewPostQueue returns new PostQueue.
func NewPostQueue(opt *PostQueueOpt) *PostQueue {
	return &PostQueue{opt: opt}
}
