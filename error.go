package zarinpal

import "fmt"

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
type Err struct {
	Code    uint8
	Message string
}

func (e Err) Error() string {
	return fmt.Sprintf("zarinpal-error %d - %s", e.Code, e.Message)
}
