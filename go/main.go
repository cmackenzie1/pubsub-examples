package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

func main() {
	u, err := url.Parse(os.Getenv("BROKER_URI"))
	if err != nil {
		panic(err)
	}

	cfg := autopaho.ClientConfig{
		BrokerUrls: []*url.URL{u},
	}

	cfg.SetUsernamePassword("", []byte(os.Getenv("BROKER_TOKEN")))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to the broker - this will return immediately after initiating the connection process
	cm, err := autopaho.NewConnection(ctx, cfg)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	// Start off a goRoutine that publishes messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count uint64
		for {
			// AwaitConnection will return immediately if connection is up; adding this call stops publication whilst
			// connection is unavailable.
			err = cm.AwaitConnection(ctx)
			if err != nil { // Should only happen when context is cancelled
				fmt.Printf("publisher done (AwaitConnection: %s)\n", err)
				return
			}

			count += 1
			// The message could be anything; lets make it JSON containing a simple count (makes it simpler to track the messages)
			msg, err := json.Marshal(struct {
				Count uint64
			}{Count: count})
			if err != nil {
				panic(err)
			}

			// Publish will block so we run it in a goRoutine
			go func(msg []byte) {
				_, err := cm.Publish(ctx, &paho.Publish{
					QoS:     byte(0),
					Topic:   os.Getenv("BROKER_TOPIC"),
					Payload: msg,
				})
				if err != nil {
					fmt.Printf("error publishing: %s\n", err)
				}
			}(msg)

			select {
			case <-time.After(1 * time.Second):
			case <-ctx.Done():
				fmt.Println("publisher done")
				return
			}
		}
	}()

	// Wait for a signal before exiting
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	fmt.Println("signal caught - exiting")
	cancel()

	wg.Wait()
	fmt.Println("shutdown complete")
}
