package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type APIResponse struct {
	Source  string
	Address map[string]interface{}
	Error   error
}

func main() {
	cep := "01153000"
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resultChan := make(chan APIResponse, 2)

	go fetchFromBrasilAPI(ctx, cep, resultChan)
	go fetchFromViaCEP(ctx, cep, resultChan)

	select {
	case <-ctx.Done():
		fmt.Println("❌ Timeout: no API responded within 1 second")
	case res := <-resultChan:
		if res.Error != nil {
			fmt.Println("❌ Error:", res.Error)
			return
		}
		fmt.Printf("✅ Fastest response from: %s\n", res.Source)
		printAddress(res.Address)
	}
}

func fetchFromBrasilAPI(ctx context.Context, cep string, ch chan<- APIResponse) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	fetch(ctx, url, "BrasilAPI", ch)
}

func fetchFromViaCEP(ctx context.Context, cep string, ch chan<- APIResponse) {
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)
	fetch(ctx, url, "ViaCEP", ch)
}

func fetch(ctx context.Context, url, source string, ch chan<- APIResponse) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ch <- APIResponse{Source: source, Error: err}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- APIResponse{Source: source, Error: err}
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- APIResponse{Source: source, Error: err}
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		ch <- APIResponse{Source: source, Error: err}
		return
	}

	ch <- APIResponse{Source: source, Address: result}
}

func printAddress(address map[string]interface{}) {
	for k, v := range address {
		fmt.Printf("%s: %v\n", k, v)
	}
}
