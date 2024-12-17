package main

import (
	"fmt"
	"pengui/reqManager"
	"pengui/utils"
	"sort"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go"
)

type WalletResult struct {
	PublicKey string
	Unclaimed int64
	Error     error
}

func main() {
	logger := utils.NewLogger()
	logger.Info("Starting the application")

	proxies := utils.SetupProxy()
	logger.Info(fmt.Sprintf("Proxies set up. Total proxies: %d", len(proxies)))

	wallets := utils.SetupWallets()
	logger.Info(fmt.Sprintf("Wallets set up. Total wallets: %d", len(wallets)))

	results := make([]WalletResult, len(wallets))
	var wg sync.WaitGroup

	logger.Info("Beginning to process wallets")

	processWallet := func(index int, wallet solana.Wallet) {
		defer wg.Done()

		logger.Info(fmt.Sprintf("Processing wallet %d: %s", index+1, wallet.PublicKey()))

		startTime := time.Now()
		res, err := reqManager.GetAllocationSize(wallet.PublicKey().String(), proxies)
		elapsedTime := time.Since(startTime)

		if err != nil {
			logger.Error(fmt.Sprintf("Error processing wallet %d: %s. Error: %s", index+1, wallet.PublicKey(), err.Error()))
			results[index] = WalletResult{PublicKey: wallet.PublicKey().String(), Error: err}
		} else {
			logger.Info(fmt.Sprintf("Result from wallet %d: %s, unclaimed: %d, processing time: %v", index+1, wallet.PublicKey(), res.TotalUnclaimed, elapsedTime))
			results[index] = WalletResult{PublicKey: wallet.PublicKey().String(), Unclaimed: int64(res.TotalUnclaimed)}
		}
	}

	for index, wallet := range wallets {
		time.Sleep(50 * time.Millisecond)
		walletObj, err := solana.WalletFromPrivateKeyBase58(wallet.String())
		if err != nil {
			logger.Error(err.Error())
		}
		wg.Add(1)
		go processWallet(index, *walletObj)
	}

	logger.Info("Waiting for all goroutines to complete")
	wg.Wait()

	// Retry failed wallets
	for index, result := range results {
		if result.Error != nil {
			logger.Info(fmt.Sprintf("Retrying failed wallet %s", result.PublicKey))
			walletObj, err := solana.WalletFromPrivateKeyBase58(wallets[index].String())
			if err != nil {
				logger.Error(fmt.Sprintf("Error creating wallet object for retry: %s", err.Error()))
				continue
			}
			wg.Add(1)
			go processWallet(index, *walletObj)
		}
	}

	wg.Wait()

	logger.Info("All goroutines completed. Generating result table.")

	// Sort results by unclaimed tokens (descending order)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Unclaimed > results[j].Unclaimed
	})

	// Generate and print the table
	fmt.Println("\nWallet Results Table:")
	fmt.Println("----------------------------------------------------")
	fmt.Printf("%-42s | %-15s\n", "Wallet Public Key", "Unclaimed Tokens")
	fmt.Println("----------------------------------------------------")
	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("%-42s | %-15s\n", result.PublicKey, "Failed")
		} else {
			fmt.Printf("%-42s | %-15d\n", result.PublicKey, result.Unclaimed)
		}
	}
	fmt.Println("----------------------------------------------------")

	logger.Info("Application finished.")
}
