/*
 * MIT License
 *
 * Copyright (c) 2022 Anton (karmadon) Stremovskyy <stremovskyy@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package gofondy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type client struct {
	httpClient *http.Client
	debug      bool
}

func newClient(debug bool, timeout, keepAlive, idleTimeout time.Duration) *client {
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: keepAlive,
	}

	tr := &http.Transport{
		IdleConnTimeout:    idleTimeout,
		DisableCompression: true,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		}}

	return &client{debug: debug, httpClient: &http.Client{Transport: tr}}
}

func (m *client) payment(ctx context.Context, url FondyURL, request *RequestObject, merchantAccount *MerchantAccount) (*[]byte, error) {
	return m.do(ctx, url, request, false, merchantAccount, true)
}

func (m *client) withdraw(ctx context.Context, url FondyURL, request *RequestObject, merchantAccount *MerchantAccount) (*[]byte, error) {
	return m.do(ctx, url, request, true, merchantAccount, true)
}

func (m *client) final(ctx context.Context, url FondyURL, request *RequestObject, merchantAccount *MerchantAccount) (*[]byte, error) {
	return m.do(ctx, url, request, false, merchantAccount, false)
}

func (m *client) do(ctx context.Context, url FondyURL, request *RequestObject, credit bool, merchantAccount *MerchantAccount, addOrderDescription bool) (*[]byte, error) {
	requestID := uuid.New().String()

	request.MerchantID = &merchantAccount.MerchantID

	if addOrderDescription {
		request.OrderDesc = StringRef(merchantAccount.MerchantString)
	}

	if credit {
		err := request.Sign(merchantAccount.MerchantCreditKey)
		if err != nil {
			return nil, fmt.Errorf("cannot sign request with credit key: %w", err)
		}
	} else {
		err := request.Sign(merchantAccount.MerchantKey)
		if err != nil {
			return nil, fmt.Errorf("cannot sign request with merchant key: %w", err)
		}
	}

	jsonValue, err := json.Marshal(NewFondyRequest(request))
	if err != nil {
		return nil, fmt.Errorf("cannot marshal request: %w", err)
	}

	if m.debug {
		fmt.Printf("Fondy request: %s \n", jsonValue)
	}

	req, err := http.NewRequest(http.MethodPost, string(url), bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("cannot create network request (url: %s): %w", url, err)
	}

	req.Header = http.Header{
		"User-Agent":   {"GOFondy/" + Version},
		"Accept":       {"application/json"},
		"Content-Type": {"application/json"},
		"X-Request-ID": {requestID},
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot do network request (url: %s): %w", url, err)
	}

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read buffer from request ID: %s (url: %s): %w", requestID, url, err)
	}

	if m.debug {
		fmt.Printf("Fondy raw response: %s \n", raw)
	}

	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot copy request ID: %s (url: %s): %w", requestID, url, err)
	}

	defer resp.Body.Close()

	return &raw, nil
}
