package stripe

type errorBody struct {
	Error stripeError `json:"error"`
}

type stripeError struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Param   string `json:"param,omitempty"`
	Message string `json:"message"`
}

type checkoutSessionResponse struct {
	ID                string `json:"id,omitempty"`
	Object            string `json:"object"`
	PaymentStatus     string `json:"payment_status"`
	ClientReferenceID string `json:"client_reference_id,omitempty"`
}

type paymentIntentResponse struct {
	ID       string `json:"id,omitempty"`
	Object   string `json:"object"`
	Status   string `json:"status"`
	Amount   int    `json:"amount,omitempty"`
	Currency string `json:"currency,omitempty"`
}

type listResponse struct {
	Object  string `json:"object"`
	Data    any    `json:"data"`
	HasMore bool   `json:"has_more"`
	URL     string `json:"url"`
}
