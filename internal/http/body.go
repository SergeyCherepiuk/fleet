package http

import (
	"encoding/json"
	"io"
	"net/http"
)

func Body(resp *http.Response, v any) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}
