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
	"crypto/sha1"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type Request struct {
	Request *RequestObject `json:"request"`
}

func NewFondyRequest(request *RequestObject) *Request {
	return &Request{Request: request}
}

// RequestObject Accept purchase (hosted payment page)
type RequestObject struct {
	OrderID    *string `json:"order_id"`
	MerchantID *string `json:"merchant_id"`
	Signature  *string `json:"signature"`

	OrderDesc         *string `json:"order_desc,omitempty"`
	Amount            *string `json:"amount,omitempty"`
	Currency          *string `json:"currency,omitempty"`
	Preauth           *string `json:"preauth,omitempty"`
	DesignID          *string `json:"design_id,omitempty"`
	Rectoken          *string `json:"rectoken,omitempty"`
	ProductID         *string `json:"product_id,omitempty"`
	Lang              *string `json:"lang,omitempty"`
	SenderEmail       *string `json:"sender_email,omitempty"`
	ServerCallbackURL *string `json:"server_callback_url,omitempty"`
	Lifetime          *string `json:"lifetime,omitempty"`
	Verification      *string `json:"verification,omitempty"`
	RequiredRectoken  *string `json:"required_rectoken,omitempty"`
	MerchantData      *string `json:"merchant_data,omitempty"`
	ReceiverRectoken  *string `json:"receiver_rectoken,omitempty"`
	Container         *string `json:"container,omitempty"`

	AdditionalData map[string]string `json:"-"`
}

// Sign - adds signature for request using provided key
func (r *RequestObject) Sign(key string) error {
	if r.Signature != nil {
		r.Signature = nil
	}

	s := key + "|"

	values := reflect.ValueOf(*r)
	types := values.Type()
	preFiltered := map[string]string{}

	for i := 0; i < values.NumField(); i++ {
		if types.Field(i).Name == "AdditionalData" {
			continue
		}

		t := values.Field(i).Interface()

		if t != nil {
			s, ok := t.(*string)
			if ok && s != nil {
				preFiltered[types.Field(i).Name] = *s
			}
		}
	}

	keys := make([]string, 0, len(preFiltered))
	for k := range preFiltered {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	final := make([]string, 0, len(preFiltered))
	for _, k := range keys {
		final = append(final, preFiltered[k])
	}

	s += strings.Join(final, "|")

	h := sha1.New()
	h.Write([]byte(s))
	r.Signature = StringRef(fmt.Sprintf("%x", h.Sum(nil)))

	return nil
}
