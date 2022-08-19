# Zarinpal API in Go
Please read [zarinpal docs](https://docs.zarinpal.com/paymentGateway/guide/) first

## 0 - Installation


```go
go get github.com/xbozorg/zarinpal-api/api/v1
```

## 1 - Zarinpal instance
```go
const (
	merchantID = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" // 36 characters
	sandbox    = true
)
// new zarinpal instance
var z = v1.New(merchantID, sandbox) 
```

## 2 - Payment Request Data
```go
PaymentRequestData := entity.PaymentRequest{
    MerchantID:  z.MerchantID,
    Amount:      110000,
    Description: "test",
    CallbackURL: "http://example.com/payment/check",
}
```

## 3 - Payment Request
```go
PaymentResponseData, err := z.PaymentRequest(
    PaymentRequestData, 
    validation.ValidatePayment(),
)
```


## 4 - Payment Gateway
Add `Payment Response Data`'s `Authority` field in previous step to `GatewayURL` and give the link to user.
- Authority : 000000000000000000000000000000111111
- GatewayURL : https://sandbox.zarinpal.com/pg/StartPay/

Payment Link : https://sandbox.zarinpal.com/pg/StartPay/000000000000000000000000000000111111

## 4 - Payment Verification Data
Get `Status` and `Authority` query parameter values at the end of your `CallbackURL`:

- CallbackURL -> https://example.com/payment/check
- Zarinpal redirects to -> https://example.com/payment/check?Authority=exampleAurhority&Status=exampleStatus

```go
verificationRequestData := entity.PaymentVerificationRequest{
    MerchantID: z.MerchantID,
    Amount:     110000,
    Authority:  "000000000000000000000000000000111111",
}
```

## 4 - Payment Verification
```go
verificationResponse, err := z.PaymentVerification(
    verificationRequestData,
    validation.ValidatePaymentVerification(),
)
```

## 5 - Check Verification Response
If `verificationResponse`'s `Status` field == 100 or 101 it means that the payment was successful.
- Status = 100 : Successful / First Verification
- Status = 101 : Successful / Already Verified

---

## Error Codes
```lua
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
```
##  Error handling
```go
if err != nil {
    if err.(v1.ZarinpalError).Code == 10 {
        // ...
    }
    if err.(v1.ZarinpalError).Code == 11 {
        // ...
    }
}
```
