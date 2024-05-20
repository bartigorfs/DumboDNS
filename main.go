package main

import (
	core "JumboDNS/core"
	"JumboDNS/database"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/miekg/dns"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func StartDNSServer() {
	var err error
	core.LocalRecords, err = core.LoadDNSRecords("temp.json")
	if err != nil {
		log.Fatalf("Error while loading persistent DNS records: %v", err)
	}

	log.Printf("Local DNS records: %s", core.LocalRecords)

	handler := new(core.DnsHandler)
	server := &dns.Server{
		Addr:      ":53",
		Net:       "udp",
		Handler:   handler,
		UDPSize:   65535,
		ReusePort: true,
	}

	log.Printf("DumboDNS started DNS server on %s", server.Addr)

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n", err.Error())
	}
}

func main() {
	gob.Register(&dns.A{})
	gob.Register(&dns.CNAME{})
	gob.Register(&dns.SOA{})
	gob.Register(&dns.PTR{})
	gob.Register(&dns.MX{})
	gob.Register(&dns.TXT{})
	gob.Register(&dns.SRV{})
	gob.Register(&dns.NS{})
	gob.Register(&dns.AAAA{})
	gob.Register(&dns.OPT{})

	ctxWc, runCancel := context.WithCancel(context.Background())

	signalWCH := make(chan os.Signal, 1)
	signal.Notify(signalWCH, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signalWCH
		log.Printf("Received signal: %v", sig)
		runCancel()
	}()

	go func() {
		core.RCC = database.RedisCacheClient()

		ctx := context.Background()
		if core.RCC == nil {
			log.Fatal("Redis client is not initialized")
		}

		err := core.RCC.Publish(ctx, "DUMBO_SUB", "test").Err()
		if err != nil {
			log.Fatalf("Error publishing message: %s", err)
		} else {
			log.Printf("Published message to %s: %s", "DUMBO_SUB", "test")
		}

		database.LocalDNSSub = core.RCC.Subscribe(ctx, "DUMBO_SUB")
		defer func(LocalDNSSub *redis.PubSub) {
			if LocalDNSSub != nil {
				err := LocalDNSSub.Close()
				if err != nil {
					log.Printf("Error closing subscription: %s", err)
				}
			}
		}(database.LocalDNSSub)

		for {
			msg, err := database.LocalDNSSub.ReceiveMessage(ctx)
			if err != nil {
				log.Fatalf("Error receiving message: %s", err)
			}

			fmt.Printf("Received message from %s: %s\n", msg.Channel, msg.Payload)
		}
	}()

	go func() {
		StartDNSServer()
	}()

	<-ctxWc.Done()
	log.Println("Shutting down...")
}
