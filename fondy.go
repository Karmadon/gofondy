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
	"context"
	"strconv"

	"github.com/google/uuid"
)

type gateway struct {
	client  *client
	options *Options
}

func New(options *Options) FondyGateway {
	c := newClient(options.Debug, options.Timeout, options.KeepAlive, options.IdleConnTimeout)
	return &gateway{options: options, client: c}
}

func (g *gateway) VerificationLink(ctx context.Context, invoiceId uuid.UUID, email *string, note string, code CurrencyCode, merchantAccount *MerchantAccount) (*string, error) {
	fondyVerificationAmount := g.options.VerificationAmount * 100
	lf := strconv.FormatFloat(g.options.VerificationLifeTime.Seconds(), 'f', 2, 64)
	cbu := g.options.CallbackBaseURL + g.options.CallbackUrl

	request := &RequestObject{
		MerchantData:      StringRef(note + "/card verification"),
		Amount:            StringRef(strconv.Itoa(fondyVerificationAmount)),
		OrderID:           StringRef(invoiceId.String()),
		OrderDesc:         StringRef(g.options.VerificationDescription),
		Lifetime:          StringRef(lf),
		Verification:      StringRef("Y"),
		DesignID:          StringRef(merchantAccount.MerchantDesignID),
		MerchantID:        StringRef(merchantAccount.MerchantID),
		RequiredRectoken:  StringRef("Y"),
		Currency:          StringRef(code.String()),
		ServerCallbackURL: StringRef(cbu),
		SenderEmail:       email,
	}

	paymentResponse, err := g.client.payment(ctx, FondyURLGetVerification, request, merchantAccount)
	if err != nil {
		return nil, NewAPIError(800, "Http request failed", err, request, paymentResponse)
	}

	fondyResponse, err := UnmarshalFondyResponse(*paymentResponse)
	if err != nil {
		return nil, NewAPIError(801, "Unmarshal response fail", err, request, paymentResponse)
	}

	if fondyResponse.Response.CheckoutURL == nil {
		return nil, NewAPIError(802, "No Url In Response", err, request, paymentResponse)
	}

	return fondyResponse.Response.CheckoutURL, nil
}
