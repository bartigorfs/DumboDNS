package main

import (
	core "JumboDNS/core"
	"JumboDNS/database"
	"context"
	"github.com/miekg/dns"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func CloseRedisSub(LocalDNSSub *redis.PubSub) {
	if LocalDNSSub != nil {
		err := LocalDNSSub.Close()
		if err != nil {
			log.Printf("Error closing subscription: %s", err)
		}
	}
}

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

		database.LocalDNSSub = core.RCC.Subscribe(ctx, "DUMBO_SUB")
		defer CloseRedisSub(database.LocalDNSSub)
		core.HandleRedisSubUpdates(ctx)
	}()

	go func() {
		StartDNSServer()
	}()

	<-ctxWc.Done()
	log.Println("Shutting down...")
}
