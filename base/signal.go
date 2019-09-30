package base

import (
	"os"
	"os/signal"
	"syscall"
)

var onlyOneSignalHandler = make(chan struct{})

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

// 监控信号退出
func SetupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler)

	stop := make(chan struct{})

	c := make(chan os.Signal, 2)

	signal.Notify(c, shutdownSignals...)

	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1)
	}()

	return stop
}
