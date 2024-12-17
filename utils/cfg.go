package utils

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"

	"time"

	"github.com/gagliardetto/solana-go"
)

// SetupWallets reads wallets from a file and returns a slice of solana.PrivateKey
func SetupWallets() []solana.PrivateKey {
	// Define the file name
	fileName := "wallets.txt"

	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Create a slice to store the wallets
	var wallets []solana.PrivateKey

	// Use a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Convert the line to a solana.PrivateKey
		privateKey := solana.MustPrivateKeyFromBase58(line)

		// Append the private key to the slice
		wallets = append(wallets, privateKey)
	}

	// Check for errors while scanning
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	return wallets
}

// SetupWallets reads wall
func SetupProxy() []*url.URL {
	// Define the file name
	fileName := "proxy.txt"

	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Create a slice to store the proxies
	var proxies []*url.URL

	// Use a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")

		if len(parts) != 4 {
			log.Printf("Invalid proxy format: %s", line)
			continue
		}

		// Construct the proxy URL
		proxyURL := &url.URL{
			Scheme: "http", // You can change this to "https" if needed
			User:   url.UserPassword(parts[2], parts[3]),
			Host:   fmt.Sprintf("%s:%s", parts[0], parts[1]),
		}

		// Append the proxy URL to the slice
		proxies = append(proxies, proxyURL)
	}

	// Check for errors while scanning
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	return proxies
}

func GetRandomProxy(proxies []*url.URL) *url.URL {
	if len(proxies) == 0 {
		return nil
	}
	rand.Seed(time.Now().UnixNano())
	randomProxy := proxies[rand.Intn(len(proxies))]
	return randomProxy

}
