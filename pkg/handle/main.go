package handle

import (
	"encoding/json"
	"errors"
	"net/http"
	"preebot/pkg/logger"
)
type (
	Handle struct {
		Hex               string            `json:"hex,omitempty"`
		Name              string            `json:"name,omitempty"`
		Image             string            `json:"image,omitempty"`
		StandardImage     string            `json:"standard_image,omitempty"`
		Holder            string            `json:"holder,omitempty"`
		HolderType        string            `json:"holder_type,omitempty"`
		Length            int               `json:"length,omitempty"`
		OgNumber          int               `json:"og_number,omitempty"`
		Rarity            string            `json:"rarity,omitempty"`
		Utxo              string            `json:"utxo,omitempty"`
		Characters        string            `json:"characters,omitempty"`
		NumericModifiers  string            `json:"numeric_modifiers,omitempty"`
		DefaultInWallet   string            `json:"default_in_wallet,omitempty"`
		ResolvedAddresses ResolvedAddresses `json:"resolved_addresses,omitempty"`
		CreatedSlotNumber int               `json:"created_slot_number,omitempty"`
		UpdatedSlotNumber int               `json:"updated_slot_number,omitempty"`
		HasDatum          bool              `json:"has_datum,omitempty"`
		ImageHash         string            `json:"image_hash,omitempty"`
		StandardImageHash string            `json:"standard_image_hash,omitempty"`
		SvgVersion        string            `json:"svg_version,omitempty"`
		LastUpdateAddress string            `json:"last_update_address,omitempty"`
		Version           int               `json:"version,omitempty"`
		HandleType        string            `json:"handle_type,omitempty"`
		PaymentKeyHash    string            `json:"payment_key_hash,omitempty"`
		PzEnabled         bool              `json:"pz_enabled,omitempty"`
		Policy            string            `json:"policy,omitempty"`
	}
	ResolvedAddresses struct {
		Ada string `json:"ada,omitempty"`
	}
)

var (
    HANDLE_API = "https://api.handle.me/"
	HANDLES = "/handles/"
)

func Lookup(uHandle string) (string, error) {
    if uHandle == "" {
        return "", errors.New("handle is empty")
    }

    url := HANDLE_API + HANDLES + uHandle
    resp, err := http.Get(url)
    if err != nil {
		logger.Record.Error("could not lookup handle", "ERROR", err)
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
		logger.Record.Error("failed to fetch handle data", "URL", url, "STATUS_CODE", resp.StatusCode, "REASON", resp.Status, "BODY", resp.Body)
        return "", errors.New("failed to fetch handle data")
    }

    var result Handle
    err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
        return "", err
    }

	if result.ResolvedAddresses.Ada == "" {
		logger.Record.Error("no ada address found in handle data", "HANDLE", uHandle)
		return "", errors.New("no ada address found in handle data")
	}

    return result.ResolvedAddresses.Ada, nil

}