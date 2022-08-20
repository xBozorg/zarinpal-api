package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"time"

	"github.com/xbozorg/zarinpal-api/config"
	"github.com/xbozorg/zarinpal-api/entity"
	"github.com/xbozorg/zarinpal-api/validation"
)

const (
	APIPaymentURL        = "https://api.zarinpal.com/pg/v4/payment/request.json"
	APIPaymentGatewayURL = "https://zarinpal.com/pg/StartPay/"
	APIVerificationURL   = "https://api.zarinpal.com/pg/v4/payment/verify.json"

	SandboxPaymentURL        = "https://sandbox.zarinpal.com/pg/rest/WebGate/PaymentRequest.json"
	SandboxPaymentGatewayURL = "https://sandbox.zarinpal.com/pg/StartPay/"
	SandboxVerificationURL   = "https://sandbox.zarinpal.com/pg/rest/WebGate/PaymentVerification.json"
)

/*
Errors :

	Code 10 -> payment validator
	Code 11 -> payment marshaling
	Code 12 -> new payment request
	Code 13 -> send payment request
	Code 14 -> read payment response body
	Code 15 -> unmarshaling payment response

	Code 30 -> verification validator
	Code 31 -> verification marshaling
	Code 32 -> new verification request
	Code 33 -> send verification request
	Code 34 -> read verification response
	Code 35 -> unmarshaling verification response
*/
type ZarinpalError struct {
	Code    uint8
	Message string
}

func (ze ZarinpalError) Error() string {
	return fmt.Sprintf("zarinpal-error %d - %s", ze.Code, ze.Message)
}

type ZarinPal struct {
	MerchantID    string
	Sandbox       bool
	DefaultConfig config.Config
}

func New(merchantID string, sandbox bool) ZarinPal {
	if sandbox {
		return ZarinPal{
			MerchantID: merchantID,
			Sandbox:    sandbox,
			DefaultConfig: config.Config{
				PaymentURL:        SandboxPaymentURL,
				PaymentGatewayURL: SandboxPaymentGatewayURL,
				VerificationURL:   SandboxVerificationURL,
			},
		}
	}

	return ZarinPal{
		MerchantID: merchantID,
		Sandbox:    sandbox,
		DefaultConfig: config.Config{
			PaymentURL:        APIPaymentURL,
			PaymentGatewayURL: APIPaymentGatewayURL,
			VerificationURL:   APIVerificationURL,
		},
	}
}

/*
	<Step 1>

I) Gets entity.PaymentRequest :

	{
	     MerchantID  string            `json:"MerchantID"`  --> Required / 36 Characters
	     Amount      uint              `json:"Amount"`      --> Required / Amount in Rial
	     Description string            `json:"Description"` --> Required
	     CallbackURL string            `json:"CallbackURL"` --> Required
	     Metadata    map[string]string `json:"metadata"`    --> Optional / {"mobile":"09111111111","email":"example@gmail.com"}
	}

II) Sends a POST request with data in I) to ZarinPal.DefaultConfig.PaymentURL --> "https://api.zarinpal.com/pg/v4/payment/request.json" or "https://sandbox.zarinpal.com/pg/rest/WebGate/PaymentRequest.json"

III) Gets a JSON response and unmarshals it to entity.PaymentResponse :

	     {
	          Status    int    `json:"Status"`                 --> 100 : request is OK
		      Authority string `json:"Authority"`              --> 36 Digits
	     }
*/
func (z ZarinPal) PaymentRequest(req entity.PaymentRequest, validator validation.ValidatePaymentRequest) (entity.PaymentResponse, error) {

	err := validator(req)

	if err != nil {
		return entity.PaymentResponse{}, ZarinpalError{
			Code:    10,
			Message: fmt.Sprintf("payment validator : %s", err.Error()),
		}
	}

	marshaledRequest, err := json.Marshal(req)
	if err != nil {
		return entity.PaymentResponse{}, ZarinpalError{
			Code:    11,
			Message: fmt.Sprintf("payment marshaling : %s", err.Error()),
		}
	}

	paymentRequest, err := http.NewRequest(
		"POST",
		z.DefaultConfig.PaymentURL,
		bytes.NewReader(marshaledRequest),
	)
	if err != nil {
		return entity.PaymentResponse{}, ZarinpalError{
			Code:    12,
			Message: fmt.Sprintf("new payment request : %s", err.Error()),
		}
	}
	paymentRequest.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 10 * time.Second}
	paymentResponse, err := client.Do(paymentRequest)
	if err != nil {
		return entity.PaymentResponse{}, ZarinpalError{
			Code:    13,
			Message: fmt.Sprintf("send payment request : %s", err.Error()),
		}
	}

	responseBytes, err := io.ReadAll(paymentResponse.Body)
	if err != nil {
		return entity.PaymentResponse{}, ZarinpalError{
			Code:    14,
			Message: fmt.Sprintf("read payment response body : %s", err.Error()),
		}
	}

	responseJSON := entity.PaymentResponse{}

	err = json.Unmarshal(responseBytes, &responseJSON)
	if err != nil {
		return entity.PaymentResponse{}, ZarinpalError{
			Code:    15,
			Message: fmt.Sprintf("unmarshaling payment response : %s", err.Error()),
		}
	}

	return responseJSON, nil
}

/* <Step 2>
I) Use Authority in Step 1 and add it to ZarinPal.DefaultConfig.PaymentGatewayURL in order to generate the payment link:

   https://zarinpal.com/pg/StartPay/    or    https://sandbox.zarinpal.com/pg/StartPay/      + Authority

II) Give payment link to user
*/

/* <Step 3>
I) When the payment confirmed/failed , zarinpal redirects to CallbackURL in <Step 1 I> with a GatewayResponse :
   {
	    Status    string `json:"Status"`    --> OK or NOK / OK --> successful , NOK --> unsuccessful or canceled
	    Authority string `json:"Authority"`
   }

II) Gets that GatewayResponse and a PaymentVerificationRequest :
    {
	    MerchantID string `json:"MerchantID"`
	    Amount     uint   `json:"Amount"`
	    Authority  string `json:"Authority"`
    }

III) Sends a POST request with data in II) to ZarinPal.DefaultConfig.VerificationURL --> "https://api.zarinpal.com/pg/v4/payment/verify.json" or "https://sandbox.zarinpal.com/pg/rest/WebGate/PaymentVerification.json"

IV) Gets a JSON response and unmarshals it to entity.PaymentVerificationResponse :
    {
	    Status   int    `json:"Status"`    --> 100 or 101 : verified
	    RefID    int    `json:"RefID"`     --> zarinpal's transaction refID
	    CardPan  string `json:"CardPan"`
	    CardHash string `json:"CardHash"`
	    FeeType  string `json:"FeeType"`
	    Fee      int    `json:"Fee"`
    }
*/

func (z ZarinPal) PaymentVerification(req entity.PaymentVerificationRequest, validator validation.ValidatePaymentVerificationRequest) (entity.PaymentVerificationResponse, error) {

	err := validator(req)
	if err != nil {
		return entity.PaymentVerificationResponse{}, ZarinpalError{
			Code:    30,
			Message: fmt.Sprintf("verification validator : %s", err.Error()),
		}
	}

	marshaledRequest, err := json.Marshal(req)
	if err != nil {
		return entity.PaymentVerificationResponse{}, ZarinpalError{
			Code:    31,
			Message: fmt.Sprintf("verification marshaling : %s", err.Error()),
		}
	}

	verificationRequest, err := http.NewRequest(
		"POST",
		z.DefaultConfig.VerificationURL,
		bytes.NewReader(marshaledRequest),
	)
	if err != nil {
		return entity.PaymentVerificationResponse{}, ZarinpalError{
			Code:    32,
			Message: fmt.Sprintf("new verification request : %s", err.Error()),
		}
	}
	verificationRequest.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 10 * time.Second}
	verificationResponse, err := client.Do(verificationRequest)
	if err != nil {
		return entity.PaymentVerificationResponse{}, ZarinpalError{
			Code:    33,
			Message: fmt.Sprintf("send verification request : %s", err.Error()),
		}
	}

	responseBytes, err := io.ReadAll(verificationResponse.Body)
	if err != nil {
		return entity.PaymentVerificationResponse{}, ZarinpalError{
			Code:    34,
			Message: fmt.Sprintf("read verification response : %s", err.Error()),
		}
	}

	responseJSON := entity.PaymentVerificationResponse{}

	err = json.Unmarshal(responseBytes, &responseJSON)
	if err != nil {
		return entity.PaymentVerificationResponse{}, ZarinpalError{
			Code:    35,
			Message: fmt.Sprintf("unmarshaling verification response : %s", err.Error()),
		}
	}

	return responseJSON, nil
}
