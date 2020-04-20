package mock

import (
	"context"
	"github.com/k-kinzal/postfix-prometheus-exporter/postfix/encoding/showq"
	"io/ioutil"
	"net"
	"path"
)

func Serve(ctx context.Context, fn ShowqMessageGenFunc) (string, []showq.Message) {
	childCtx, _ := context.WithCancel(ctx)

	dir, _ := ioutil.TempDir("", "")
	showqPath := path.Join(dir, "showq")

	messages := fn()
	var buf []byte
	for _, message := range messages {
		buf = append(buf, message.Bytes()...)
	}

	listen, _ := net.Listen("unix", showqPath)
	go func() {
		for {
			conn, err := listen.Accept()
			if err != nil {
				panic(err)
			}

			conn.Write(append(buf, 0))
			conn.Close()

			select {
			case <-childCtx.Done():
				listen.Close()
				return
			}
		}
	}()

	return showqPath, messages
}