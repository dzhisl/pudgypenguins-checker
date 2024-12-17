package reqManager

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"

	"pengui/utils"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

type AllocationSize struct {
	Total          int           `json:"total"`
	TotalUnclaimed int           `json:"totalUnclaimed"`
	Categories     []interface{} `json:"categories"`
	Addresses      []string      `json:"addresses"`
}

type GetTxRes []struct {
	Data       string   `json:"data"`
	Signatures []string `json:"signatures"`
}

func createClient(proxies []*url.URL) (tls_client.HttpClient, error) {
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(10),
		tls_client.WithClientProfile(profiles.Firefox_133),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	if len(proxies) > 0 {
		randProxy := utils.GetRandomProxy(proxies)
		client.SetProxy(randProxy.String())
	}

	return client, nil
}

func setCommonHeaders(req *http.Request) {
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("origin", "https://claim.pudgypenguins.com")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://claim.pudgypenguins.com/")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "cross-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
}

func makeRequest(client tls_client.HttpClient, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	setCommonHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("Status code: %d\n", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func GetAllocationSize(walletPubKey string, proxies []*url.URL) (*AllocationSize, error) {
	client, err := createClient(proxies)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.clusters.xyz/v0.1/airdrops/pengu/eligibility/%s", walletPubKey)
	body, err := makeRequest(client, url)
	if err != nil {
		return nil, err
	}

	var result AllocationSize
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	log.Printf("Allocation Size: %+v\n", result) // Log the result for debugging
	return &result, nil
}
