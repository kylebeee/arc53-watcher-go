package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kylebeee/arc53-watcher-go/server"
)

const (
	// TimeFormat is the server format for time.Now().Format()
	TimeFormat = "2006-01-02 15:04:05"
	// TimeZone is the server time zone
	TimeZone = "America/Los_Angeles"
)

func main() {
	s := server.New()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigc
		fmt.Printf("[SERVER][%s] Shutting Down\n", time.Now().Format(TimeFormat))
		// close db
		s.Close()
		// cancel watcher
		s.WatcherCancelFn()
	}()

	s.Run(":3000")
}
