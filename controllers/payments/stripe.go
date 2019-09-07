/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package payments

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
	"net/http"
	"os"
)

// StripeWebHookHandler handles stripe webhooks.
func StripeWebHookHandler(ctx echo.Context) error {
	bodyStream, err := ctx.Request().GetBody()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(bodyStream)
	if err != nil {
		log.Error(err)
	}

	event, err := webhook.ConstructEvent(buf.Bytes(), ctx.Request().Header.Get("Stripe-Signature"), os.Getenv("STRIPE_WEBHOOK_SECRET"))

	if err != nil {
		log.Errorf("Error verifying stripe webhook signature: %v\n", err)
		return ctx.NoContent(http.StatusBadRequest) // Return a 400 error on a bad signature
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			log.Errorf("Error parsing stripe webhook JSON: %v\n", err)
			return ctx.NoContent(http.StatusBadRequest)
		}
		handlePaymentIntentSucceeded(paymentIntent)
	case "payment_method.attached":
		var paymentMethod stripe.PaymentMethod
		err := json.Unmarshal(event.Data.Raw, &paymentMethod)
		if err != nil {
			log.Errorf("Error parsing stripe webhook JSON: %v\n", err)
			return ctx.NoContent(http.StatusBadRequest)
		}
		handlePaymentMethodAttached(paymentMethod)
	default:
		log.Errorf("Unexpected stripe webhook event type: %s\n", event.Type)
		return ctx.NoContent(http.StatusBadRequest)
	}

	return ctx.JSON(http.StatusOK, "")
}

func handlePaymentMethodAttached(method stripe.PaymentMethod) {

}

func handlePaymentIntentSucceeded(intent stripe.PaymentIntent) {

}
