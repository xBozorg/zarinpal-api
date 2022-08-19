package entity

type PaymentRequest struct {
	MerchantID  string            `json:"MerchantID"`
	Amount      int               `json:"Amount"`
	Description string            `json:"Description"`
	CallbackURL string            `json:"CallbackURL"`
	Metadata    map[string]string `json:"Metadata"`
}
type PaymentResponse struct {
	Status    int    `json:"Status"`
	Authority string `json:"Authority"`
}

type GatewayResponse struct {
	Status    string `json:"Status"`
	Authority string `json:"Authority"`
}

type PaymentVerificationRequest struct {
	MerchantID string `json:"MerchantID"`
	Amount     int    `json:"Amount"`
	Authority  string `json:"Authority"`
}
type PaymentVerificationResponse struct {
	Status   int    `json:"Status"`
	RefID    int    `json:"RefID"`
	CardPan  string `json:"CardPan"`
	CardHash string `json:"CardHash"`
	FeeType  string `json:"FeeType"`
	Fee      int    `json:"Fee"`
}
