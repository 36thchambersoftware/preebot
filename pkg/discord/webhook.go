package discord

import (
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/exp/slog"
)

func Webhook() {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// TODO replace with webhook params struct
	data := strings.NewReader("")

	// Create a new request
	request, err := http.NewRequest(http.MethodPost, DISCORD_WEBHOOK_URL, data)
	if err != nil {
		slog.Error("PREEBOT", "PACKAGE", "DISCORD", "request", err)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		slog.Error("PREEBOT", "PACKAGE", "DISCORD", "response", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNoContent {
		slog.Warn("PREEBOT", "PACKAGE", "DISCORD", "INFO", "You are not waiting for a response. Add ?wait=true to webhook url")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("PREEBOT", "PACKAGE", "DISCORD", "body", body, "error", err)
		return
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		slog.Error("PREEBOT", "PACKAGE", "DISCORD", "STATUS", response.StatusCode, "ERROR", string(body))
		return
	}
}
