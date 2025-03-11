package taptools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"preebot/pkg/logger"
	"strings"
	"time"
)

type (
	Trades []Trade
	Trade struct {
		Action       string `json:"action,omitempty"`
		Address      string `json:"address,omitempty"`
		Exchange     string `json:"exchange,omitempty"`
		Hash         string `json:"hash,omitempty"`
		LpTokenUnit  string `json:"lpTokenUnit,omitempty"`
		Price        float64`json:"price,omitempty"`
		Time         int    `json:"time,omitempty"`
		TokenA       string `json:"tokenA,omitempty"`
		TokenAAmount float64`json:"tokenAAmount,omitempty"`
		TokenAName   string `json:"tokenAName,omitempty"`
		TokenB       string `json:"tokenB,omitempty"`
		TokenBAmount float64`json:"tokenBAmount,omitempty"`
		TokenBName   string `json:"tokenBName,omitempty"`
	}
)

var (
	TAPTOOLS_API_BASE_URL = "https://openapi.taptools.io/api/v1/"
	TAPTOOLS_API_KEY string
)

func init() {
	TAPTOOLS_API_KEY = loadTaptoolsAPIKey()
}

func loadTaptoolsAPIKey() string {
	TAPTOOLS_API_KEY, ok := os.LookupEnv("TAPTOOLS_API_KEY")
	if !ok {
		logger.Record.Error("Could not get taptools api key")
	}

	return TAPTOOLS_API_KEY
}

func GetTrades(ctx context.Context, policyAsset string) (Trades, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	data := strings.NewReader(fmt.Sprintf(`{"unit": "%s"}`, policyAsset))
	url, err := url.Parse(fmt.Sprintf("%stoken/trades?unit=%s&sortBy=time&minAmount=1&order=desc", TAPTOOLS_API_BASE_URL, policyAsset))
	if err != nil {
		return nil, err
	}
	logger.Record.Info("get trades", "URL", url.String())
	req, err := http.NewRequest(http.MethodGet, url.String(), data)
	if err != nil {
		logger.Record.Error("Could not connect to taptools api", "ERROR", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", TAPTOOLS_API_KEY)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Record.Error("invalid response from taptools api", "ERROR", err)
	}
    defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var trades Trades
	if err := json.Unmarshal(body, &trades); err != nil {
        logger.Record.Error("could not unmarshal response", "BODY", string(body), "ERROR", err)
    }
	
	return trades, nil
}
