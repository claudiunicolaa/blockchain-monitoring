package monitor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
)

func (bm *BlockchainMonitor) MonitorBitcoin(ctx context.Context) error {
	// Connect to Blockdaemon Bitcoin node
	connCfg := &rpcclient.ConnConfig{
		Host:         bm.config.BTCEndpoint,
		User:         "blockdaemon",
		Pass:         bm.config.APIKey,
		HTTPPostMode: true,
		DisableTLS:   false,
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Bitcoin node: %v", err)
	}
	defer client.Shutdown()

	if err := client.Ping(); err != nil {
		return fmt.Errorf("failed to ping Bitcoin node: %v", err)
	}

	log.Printf("Started monitoring Bitcoin blockchain")

	count, err := client.GetBlockCount()
	fmt.Printf("Block count: %d\n", count)
	if err != nil {
		return fmt.Errorf("failed to get block count: %v", err)
	}

	// Track last processed block
	bestBlock, err := client.GetBestBlockHash()
	fmt.Printf("Best block: %s\n", bestBlock)
	if err != nil {
		return fmt.Errorf("failed to get best block: %v", err)
	}

	lastBlockHash := bestBlock

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Get current best block
			currentBest, err := client.GetBestBlockHash()
			if err != nil {
				log.Printf("Error getting best block: %v", err)
				time.Sleep(time.Second)
				continue
			}

			if currentBest.String() != lastBlockHash.String() {
				block, err := client.GetBlock(currentBest)
				if err != nil {
					log.Printf("Error getting block: %v", err)
					time.Sleep(time.Second)
					continue
				}

				bm.mutex.RLock()
				for _, tx := range block.Transactions {
					for userId, addresses := range bm.addresses {
						// Check inputs and outputs for monitored addresses
						for _, input := range tx.TxIn {
							prevOut := input.PreviousOutPoint.String()
							if addresses[prevOut] {
								bm.processTransaction(Transaction{
									UserID:    userId,
									Chain:     "BTC",
									Source:    prevOut,
									Timestamp: block.Header.Timestamp,
									TxHash:    tx.TxHash().String(),
									BlockNum:  uint64(block.Header.Nonce),
								})
							}
						}
					}
				}
				bm.mutex.RUnlock()

				lastBlockHash = currentBest
			}

			// Small delay to prevent too frequent polling
			time.Sleep(time.Second)
		}
	}
}
