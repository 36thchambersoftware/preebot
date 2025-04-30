package cardano

import (
	"encoding/hex"
	"encoding/json"
	"errors"
)

type KoiosDatum struct {
    Fields []struct {
        Map []struct {
            K map[string]string `json:"k"`
            V map[string]string `json:"v"`
        } `json:"map"`
    } `json:"fields"`
}


// ParseDatumValueFixed parses the more complex Koios Datum.Value structure
func ParseDatumValueFixed(datumValue json.RawMessage) (map[string]string, error) {
    var parsed KoiosDatum
    err := json.Unmarshal(datumValue, &parsed)
    if err != nil {
        return nil, err
    }

    result := make(map[string]string)

    if len(parsed.Fields) == 0 {
        return nil, errors.New("no fields in datum")
    }

    for _, entry := range parsed.Fields[0].Map {
        keyBytesHex, keyOK := entry.K["bytes"]
        valBytesHex, valOK := entry.V["bytes"]

        if !keyOK || !valOK {
            continue
        }

        keyBytes, err := hex.DecodeString(keyBytesHex)
        if err != nil {
            continue
        }

        valBytes, err := hex.DecodeString(valBytesHex)
        if err != nil {
            continue
        }

        result[string(keyBytes)] = string(valBytes)
    }

    return result, nil
}
