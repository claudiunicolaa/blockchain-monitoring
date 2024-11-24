package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

type Monitor interface {
	// AddAddresses registers addresses for monitoring
	AddAddresses(userId string, btcAddr, ethAddr, solAddr string)
	// MonitorEthereum monitors the Ethereum blockchain for specific addresses transactions
	MonitorEthereum(ctx context.Context) error
	// MonitorSolana monitors the Solana blockchain for specific addresses transactions
	MonitorSolana(ctx context.Context) error
	// MonitorBitcoin monitors the Bitcoin blockchain for specific addresses transactions
	MonitorBitcoin(ctx context.Context) error
}

type Config struct {
	BTCEndpoint string
	ETHEndpoint string
	SOLEndpoint string
	APIKey      string
}

type Transaction struct {
	UserID      string    `json:"userId"`
	Chain       string    `json:"chain"`
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	Amount      string    `json:"amount"`
	Fees        string    `json:"fees"`
	Timestamp   time.Time `json:"timestamp"`
	TxHash      string    `json:"txHash"`
	BlockNum    uint64    `json:"blockNum"`
}

type Event struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Amount      string `json:"amount"`
	Fees        string `json:"fees"`
}

type BlockchainMonitor struct {
	config    Config
	addresses map[string]map[string]bool // userId -> addresses map
	mutex     sync.RWMutex
}

func New(config Config) *BlockchainMonitor {
	return &BlockchainMonitor{
		config:    config,
		addresses: make(map[string]map[string]bool),
	}
}

func (bm *BlockchainMonitor) AddAddresses(userId string, btcAddr, ethAddr, solAddr string) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	bm.addresses[userId] = map[string]bool{
		btcAddr: true,
		ethAddr: true,
		solAddr: true,
	}
}

func (bm *BlockchainMonitor) processTransaction(tx Transaction) {
	e := Event{
		Source:      tx.Source,
		Destination: tx.Destination,
		Amount:      tx.Amount,
		Fees:        tx.Fees,
	}

	encoded, err := json.Marshal(e)
	if err != nil {
		log.Printf("failed to marshal event: %v", err)
		return
	}

	msg := &sarama.ProducerMessage{
		Topic: "transactions",
		Key:   sarama.StringEncoder(tx.UserID),
		Value: sarama.ByteEncoder(encoded),
	}

	fmt.Println(msg)
}
