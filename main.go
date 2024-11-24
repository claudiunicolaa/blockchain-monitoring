package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/claudiunicolaa/blockchain-monitoring/internal/monitor"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	cfg := monitor.Config{
		APIKey:      os.Getenv("BLOCKDAEMON_API_KEY"),
		ETHEndpoint: os.Getenv("BLOCKDAEMON_ETH_ENDPOINT_WS"),
	}
	monitor := monitor.New(cfg)

	monitor.AddAddresses(
		"1",
		"bc1q23456789012345678901234567890123456789",
		"0x1234567890123456789012345678901234567890",
		"So111111111111111111111111111111111111111",
	)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	var waitChan = make(chan struct{})

	outEthereum := make(chan error)
	go func() {
		for err := range outEthereum {
			log.Printf("ethereum monitor error: %v", err)
		}
	}()

	go func() {
		log.Printf("Started monitoring Ethereum blockchain")
		if err := monitor.MonitorEthereum(context.Background(), outEthereum); err != nil {
			panic(err)
		}
	}()

	select {
	case <-waitChan:
		log.Println("Wait channel closed, shutting down...")
	case <-signalChan:
		log.Println("Received interrupt signal, shutting down...")
	}
	close(outEthereum)
}
