# example
```
package main

import (
	"context"
	su "github.com/ppzxc/go-prom-pprof-utils"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"time"
)

func main() {
	ctx, _ := context.WithCancel(context.Background())

	addr := net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 8080,
	}

	logger := logrus.Logger{
		Out:          os.Stdout,
		Formatter:    &logrus.TextFormatter{},
		ReportCaller: true,
		Level:        logrus.TraceLevel,
	}

	err := su.NewStatUtils(ctx, "TEST", true, &addr, &logger)
	if err != nil {
		panic(err)
	}

	su.AddGauge("fill")
	su.AddGauge("session", "type")
	su.AddCounter("fill2")
	su.AddCounter("session2", "type")
	su.AddHistogram("fill3", []float64{0.01, 0.02, 0.04, 0.08, 0.16, 0.32, 0.5, 1, 5, 10, 20, 30})
	su.AddHistogram("session3", []float64{0.01, 0.02, 0.04, 0.08, 0.16, 0.32, 0.5, 1, 5, 10, 20, 30}, "type", "query", "dd")

	go func() {
		for {
			now := time.Now()
			su.Increase("fill")
			su.Increase("session", "tcp")
			su.Increase("session", "websocket")
			su.Increase("fill2")
			su.Increase("session2", "tcp")
			su.Increase("session2", "websocket")
			su.IncreaseHistogram("fill3", now)
			su.IncreaseHistogram("session3", now, "db", "query1", "ee")
			su.IncreaseHistogram("session3", now, "db", "query3", "ee")
			su.IncreaseHistogram("session3", now, "db", "query2", "ee")
			su.IncreaseHistogram("session3", now, "db", "query5", "ee")
			time.Sleep(1 * time.Second)
		}
	}()

	su.Serve()
}
```