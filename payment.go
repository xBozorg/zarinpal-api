package zarinpal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"time"

	"github.com/xbozorg/zarinpal-api/config"
)

const (
	APIPaymentURL        = "https://api.zarinpal.com/pg/v4/payment/request.json"
	APIPaymentGatewayURL = "https://zarinpal.com/pg/StartPay/"
	APIVerificationURL   = "https://api.zarinpal.com/pg/v4/payment/verify.json"

	SandboxPaymentURL        = "https://sandbox.zarinpal.com/pg/rest/WebGate/PaymentRequest.json"
	SandboxPaymentGatewayURL = "https://sandbox.zarinpal.com/pg/StartPay/"
	SandboxVerificationURL   = "https://sandbox.zarinpal.com/pg/rest/WebGate/PaymentVerification.json"
)

type ZarinPal struct {
	MerchantID    string
	Sandbox       bool
	DefaultConfig config.Config
}

func New(merchantID string, sandbox bool) *ZarinPal {
	if sandbox {
		return &ZarinPal{
			MerchantID: merchantID,
			Sandbox:    sandbox,
			DefaultConfig: config.Config{
				PaymentURL:        SandboxPaymentURL,
				PaymentGatewayURL: SandboxPaymentGatewayURL,
				VerificationURL:   SandboxVerificationURL,
			},
		}
	}

	return &ZarinPal{
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

I) Gets PaymentRequest :

	{
	     MerchantID  string            `json:"MerchantID"`  --> Required / 36 Characters
	     Amount      uint              `json:"Amount"`      --> Required / Amount in Rial
	     Description string            `json:"Description"` --> Required
	     CallbackURL string            `json:"CallbackURL"` --> Required
	     Metadata    map[string]string `json:"metadata"`    --> Optional / {"mobile":"09111111111","email":"example@gmail.com"}
	}

II) Sends a POST request with data in I) to ZarinPal.DefaultConfig.PaymentURL --> "https://api.zarinpal.com/pg/v4/payment/request.json" or "https://sandbox.zarinpal.com/pg/rest/WebGate/PaymentRequest.json"

III) Gets a JSON response and unmarshals it to PaymentResponse :

	     {
	          Status    int    `json:"Status"`                 --> 100 : request is OK
		      Authority string `json:"Authority"`              --> 36 Digits
	     }
*/
func (z ZarinPal) PaymentRequest(req PaymentRequest, validator ValidatePaymentRequest) (PaymentResponse, error) {

	err := validator(req)

	if err != nil {
		return PaymentResponse{}, Err{
			Code:    10,
			Message: fmt.Sprintf("payment validator : %s", err.Error()),
		}
	}

	marshaledRequest, err := json.Marshal(req)
	if err != nil {
		return PaymentResponse{}, Err{
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
		return PaymentResponse{}, Err{
			Code:    12,
			Message: fmt.Sprintf("new payment request : %s", err.Error()),
		}
	}
	paymentRequest.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 10 * time.Second}
	paymentResponse, err := client.Do(paymentRequest)
	if err != nil {
		return PaymentResponse{}, Err{
			Code:    13,
			Message: fmt.Sprintf("send payment request : %s", err.Error()),
		}
	}
	defer paymentResponse.Body.Close()

	responseBytes, err := io.ReadAll(paymentResponse.Body)
	if err != nil {
		return PaymentResponse{}, Err{
			Code:    14,
			Message: fmt.Sprintf("read payment response body : %s", err.Error()),
		}
	}

	responseJSON := PaymentResponse{}

	err = json.Unmarshal(responseBytes, &responseJSON)
	if err != nil {
		return PaymentResponse{}, Err{
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

IV) Gets a JSON response and unmarshals it to PaymentVerificationResponse :
    {
	    Status   int    `json:"Status"`    --> 100 or 101 : verified
	    RefID    int    `json:"RefID"`     --> zarinpal's transaction refID
	    CardPan  string `json:"CardPan"`
	    CardHash string `json:"CardHash"`
	    FeeType  string `json:"FeeType"`
	    Fee      int    `json:"Fee"`
    }
*/

func (z ZarinPal) PaymentVerification(req PaymentVerificationRequest, validator ValidatePaymentVerificationRequest) (PaymentVerificationResponse, error) {

	err := validator(req)
	if err != nil {
		return PaymentVerificationResponse{}, Err{
			Code:    30,
			Message: fmt.Sprintf("verification validator : %s", err.Error()),
		}
	}

	marshaledRequest, err := json.Marshal(req)
	if err != nil {
		return PaymentVerificationResponse{}, Err{
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
		return PaymentVerificationResponse{}, Err{
			Code:    32,
			Message: fmt.Sprintf("new verification request : %s", err.Error()),
		}
	}
	verificationRequest.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 10 * time.Second}
	verificationResponse, err := client.Do(verificationRequest)
	if err != nil {
		return PaymentVerificationResponse{}, Err{
			Code:    33,
			Message: fmt.Sprintf("send verification request : %s", err.Error()),
		}
	}
	defer verificationResponse.Body.Close()

	responseBytes, err := io.ReadAll(verificationResponse.Body)
	if err != nil {
		return PaymentVerificationResponse{}, Err{
			Code:    34,
			Message: fmt.Sprintf("read verification response : %s", err.Error()),
		}
	}

	responseJSON := PaymentVerificationResponse{}

	err = json.Unmarshal(responseBytes, &responseJSON)
	if err != nil {
		return PaymentVerificationResponse{}, Err{
			Code:    35,
			Message: fmt.Sprintf("unmarshaling verification response : %s", err.Error()),
		}
	}

	return responseJSON, nil
}
