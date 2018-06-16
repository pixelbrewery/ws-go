//  Copyright © 2018 Pixel Brewery Co. All rights reserved.

package waysense

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type httpWriter struct {
	client    *http.Client
	url       string
	apiKey    string
	apiSecret string
}

type HttpResponse struct {
	Code   int    `json:"code"`
	Result string `json:"result"`
}

var (
	// TLSDialTimeout is the maximum amount of time a dial will wait for a connect
	// to complete.
	TLSDialTimeout      = 20 * time.Second
	TLSHandshakeTimeout = 10 * time.Second
	// HTTPClientTimeout specifies a time limit for requests made by the
	// HTTPClient. The timeout includes connection time, any redirects,
	// and reading the response body.
	HTTPClientTimeout = 60 * time.Second
	// TCPKeepAlive specifies the keep-alive period for an active network
	// connection. If zero, keep-alives are not enabled.
	TCPKeepAlive = 60 * time.Second

	// TCPIdleTimeOut is the maximum amount of time an idle
	// (keep-alive) connection will remain idle before closing
	// itself.
	// Zero means no limit.
	TCPIdleTimeOut = 0 * time.Second
)

// TODO might want to change this to udp
// timeout in duration form like 1s, 1m, 1h
func newHttpWriter(addr, apiKey, apiSecret, timeout string, skipSSL bool) (*httpWriter, error) {
	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   TLSDialTimeout,
			KeepAlive: TCPKeepAlive,
		}).Dial,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: skipSSL},
		TLSHandshakeTimeout: TLSHandshakeTimeout,
		IdleConnTimeout:     TCPIdleTimeOut,
	}

	var to time.Duration
	var err error
	if timeout != "" {
		to, err = time.ParseDuration(timeout)
		if err != nil {
			to, _ = time.ParseDuration("20s")
		}
	} else {
		to, _ = time.ParseDuration("20s")
	}

	httpClient := &http.Client{
		Timeout:   to,
		Transport: tr,
	}

	if addr == "" {
		return nil, fmt.Errorf("You must provide API url!")
	}

	return &httpWriter{
		client:    httpClient,
		url:       addr,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}, nil
}

// write to http api writer
// TODO compress body
func (w *httpWriter) Write(data []byte) (int, error) {
	var (
		err error
		req *http.Request
	)

	req, err = http.NewRequest("POST", w.url, bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	req.Header.Set("x-waysense-api-key", w.apiKey)
	req.Header.Set("x-waysense-api-secret", w.apiSecret)

	resp, err := w.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)

		result := HttpResponse{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return 0, fmt.Errorf("Bad code:%d response: %s", resp.StatusCode, string(body))
		}

		return 0, nil
	} else {
		body, _ := ioutil.ReadAll(resp.Body)

		result := HttpResponse{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return 0, fmt.Errorf("Bad code:%d response: %s", resp.StatusCode, string(body))
		}

		return 0, fmt.Errorf("%s with code:%d", result.Result, result.Code)
	}
}

func (w *httpWriter) SetWriteTimeout(d time.Duration) error {
	return nil
}

func (w *httpWriter) Close() error {
	return nil
}
