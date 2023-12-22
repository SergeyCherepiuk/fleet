package httpclient

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
)

func Get(addr, endpoint string) (*http.Response, error) {
	url, err := url.JoinPath("http://", addr, endpoint)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	return client.Get(url)
}

func Post(addr, endpoint string, payload any) (*http.Response, error) {
	url, err := url.JoinPath("http://", addr, endpoint)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	return client.Post(url, "application/json", bytes.NewReader(body))
}

func Put(addr, endpoint string, payload any) (*http.Response, error) {
	return named("PUT", addr, endpoint, payload)
}

func Patch(addr, endpoint string, payload any) (*http.Response, error) {
	return named("PATCH", addr, endpoint, payload)
}

func Delete(addr, endpoint string, payload any) (*http.Response, error) {
	return named("DELETE", addr, endpoint, payload)
}

func named(method, addr, endpoint string, payload any) (*http.Response, error) {
	url, err := url.JoinPath("http://", addr, endpoint)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	return client.Do(req)
}
