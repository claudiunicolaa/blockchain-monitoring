package monitor

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func (bm *BlockchainMonitor) MonitorEthereum(ctx context.Context, out chan error) error {
	url := fmt.Sprintf("%s?apiKey=%s", bm.config.ETHEndpoint, bm.config.APIKey)
	client, err := ethclient.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to Ethereum node: %v", err)
	}
	defer client.Close()

	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(ctx, headers)
	if err != nil {
		return fmt.Errorf("failed to subscribe to new headers: %v", err)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-sub.Err():
			return fmt.Errorf("ethereum subscription error: %v", err)
		case header := <-headers:
			block, err := client.BlockByHash(ctx, header.Hash())
			if err != nil {
				out <- fmt.Errorf("error getting block: %v", err)
				continue
			}

			bm.mutex.RLock()
			for _, tx := range block.Transactions() {
				if tx.To() == nil {
					out <- fmt.Errorf("transaction to is nil")
					continue
				}

				from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
				if err != nil {
					out <- fmt.Errorf("error getting sender: %v", err)
					continue
				}

				to := tx.To().Hex()
				for userId, addresses := range bm.addresses {
					if addresses[from.Hex()] || addresses[to] {
						bm.processTransaction(Transaction{
							UserID:      userId,
							Chain:       "ETH",
							Source:      from.Hex(),
							Destination: to,
							Amount:      tx.Value().String(),
							Fees:        new(big.Int).Mul(big.NewInt(int64(tx.Gas())), tx.GasPrice()).String(),
							Timestamp:   time.Unix(int64(block.Time()), 0),
							TxHash:      tx.Hash().Hex(),
							BlockNum:    block.NumberU64(),
						})
					}
				}
			}
			bm.mutex.RUnlock()
		}
	}
}
