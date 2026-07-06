package line

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/adapter/httpx"
)

func (r *routes) writeNotificationToken(w http.ResponseWriter, req *http.Request) {
	scenario, ok := r.resolveScenario(w, req)
	if !ok {
		return
	}
	switch scenario {
	case "auth_error":
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated invalid MINI App channel token")
	case "invalid_request":
		writeLINEError(w, http.StatusBadRequest, "[liffAccessToken] must not be blank")
	default:
		payload, err := decodePayload(req)
		if err != nil {
			writeDecodeError(w, err)
			return
		}
		if len(payload) > 0 && strings.TrimSpace(fmt.Sprint(payload["liffAccessToken"])) == "" {
			writeLINEError(w, http.StatusBadRequest, "[liffAccessToken] must not be blank")
			return
		}
		resource, err := r.store.Create("line", "notification_token", map[string]any{
			"remaining_count": 5,
			"session_id":      "line_service_session_mockport",
		})
		if err != nil {
			writeLINEError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeNotification(w, resource.ID, 5)
	}
}

func (r *routes) writeServiceMessage(w http.ResponseWriter, req *http.Request) {
	scenario, ok := r.resolveScenario(w, req)
	if !ok {
		return
	}
	switch scenario {
	case "auth_error":
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated invalid MINI App channel token")
	case "invalid_request":
		writeLINEError(w, http.StatusBadRequest, "templateName is required")
	default:
		payload, err := decodePayload(req)
		if err != nil {
			writeDecodeError(w, err)
			return
		}
		token, _ := payload["notificationToken"].(string)
		if token != "" {
			if _, ok := r.store.Get("line", "notification_token", token); !ok {
				writeLINEError(w, http.StatusBadRequest, "notificationToken is invalid")
				return
			}
		}
		resource, err := r.store.Create("line", "notification_token", map[string]any{
			"remaining_count": 4,
			"session_id":      "line_service_session_mockport",
		})
		if err != nil {
			writeLINEError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeNotification(w, resource.ID, 4)
	}
}

func (r *routes) writePayRequest(w http.ResponseWriter, req *http.Request) {
	scenario, ok := r.resolveScenarioPay(w, req)
	if !ok {
		return
	}
	switch scenario {
	case "auth_error":
		writePayError(w, "1104", "Mockport simulated LINE Pay authorization error")
	case "pay_failed":
		writePayError(w, "1169", "LINE Pay requires payment method selection and password authentication.")
	case "invalid_request":
		writePayError(w, "2101", "Mockport simulated invalid LINE Pay request")
	default:
		payload, err := decodePayload(req)
		if err != nil {
			writePayError(w, "2101", "Request body must be JSON")
			return
		}
		orderID := fmt.Sprint(firstNonEmpty(payload["orderId"], "order_mockport"))
		amount := firstNonEmpty(payload["amount"], 1000)
		currency := fmt.Sprint(firstNonEmpty(payload["currency"], "JPY"))
		resource, err := r.store.Create("line", "line_pay_payment", map[string]any{
			"orderId":  orderID,
			"amount":   amount,
			"currency": currency,
			"status":   "reserved",
		})
		if err != nil {
			writePayError(w, "9000", err.Error())
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"returnCode":    "0000",
			"returnMessage": "Success.",
			"info": map[string]any{
				"transactionId":      resource.ID,
				"paymentAccessToken": "mockport_line_pay_access_token",
				"paymentUrl": map[string]string{
					"web": adapter.LocalBaseURL(r.basePath + "/line-pay/authorize/" + resource.ID),
					"app": "line://pay/mockport/" + resource.ID,
				},
			},
		})
	}
}

func (r *routes) writePayConfirm(w http.ResponseWriter, req *http.Request, id string) {
	scenario, ok := r.resolveScenarioPay(w, req)
	if !ok {
		return
	}
	switch scenario {
	case "auth_error":
		writePayError(w, "1104", "Mockport simulated LINE Pay authorization error")
	case "pay_failed":
		writePayError(w, "1169", "LINE Pay requires payment method selection and password authentication.")
	default:
		resource, ok := r.store.Get("line", "line_pay_payment", id)
		if !ok {
			writePayError(w, "1150", "Transaction not found")
			return
		}
		updated, err := r.store.Update("line", "line_pay_payment", id, map[string]any{"status": "confirmed"})
		if err != nil {
			writePayError(w, "9000", err.Error())
			return
		}
		httpx.WriteJSON(w, http.StatusOK, payInfoResponse("0000", "Success.", updated.ID, resource.Data["orderId"], updated.Data["amount"], updated.Data["currency"], updated.Data["status"]))
	}
}

func (r *routes) writePayCheck(w http.ResponseWriter, id string) {
	resource, ok := r.store.Get("line", "line_pay_payment", id)
	if !ok {
		writePayError(w, "1150", "Transaction not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, payInfoResponse("0000", "Success.", resource.ID, resource.Data["orderId"], resource.Data["amount"], resource.Data["currency"], resource.Data["status"]))
}

func (r *routes) writeMiniDappWalletSession(w http.ResponseWriter, req *http.Request) {
	scenario, ok := r.resolveScenario(w, req)
	if !ok {
		return
	}
	if scenario == "auth_error" {
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated Mini Dapp client authorization error")
		return
	}
	if scenario == "invalid_request" {
		writeLINEError(w, http.StatusBadRequest, "chainId is required")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"sessionId":     "line_mini_dapp_wallet_session_mockport",
		"chainId":       1001,
		"walletAddress": "0x0000000000000000000000000000000000431010",
		"status":        "connected",
	})
}

func (r *routes) writeMiniDappPayment(w http.ResponseWriter, req *http.Request) {
	scenario, ok := r.resolveScenario(w, req)
	if !ok {
		return
	}
	switch scenario {
	case "auth_error":
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated Mini Dapp client authorization error")
	case "pay_failed":
		writeLINEError(w, http.StatusPaymentRequired, "Mockport simulated Mini Dapp payment failure")
	case "invalid_request":
		writeLINEError(w, http.StatusBadRequest, "itemId is required")
	default:
		payload, err := decodePayload(req)
		if err != nil {
			writeDecodeError(w, err)
			return
		}
		resource, err := r.store.Create("line", "mini_dapp_payment", map[string]any{
			"itemId":   firstNonEmpty(payload["itemId"], "item_mockport"),
			"amount":   firstNonEmpty(payload["amount"], "10"),
			"currency": firstNonEmpty(payload["currency"], "KAIA"),
			"status":   "approved",
		})
		if err != nil {
			writeLINEError(w, http.StatusInternalServerError, err.Error())
			return
		}
		body := resource.Data
		body["id"] = resource.ID
		body["checkoutUrl"] = adapter.LocalBaseURL(r.basePath + "/mini-dapp/checkout/" + resource.ID)
		httpx.WriteJSON(w, http.StatusOK, body)
	}
}

func (r *routes) writeMiniDappPaymentLookup(w http.ResponseWriter, id string) {
	resource, ok := r.store.Get("line", "mini_dapp_payment", id)
	if !ok {
		writeLINEError(w, http.StatusNotFound, "Mini Dapp payment not found")
		return
	}
	body := resource.Data
	body["id"] = resource.ID
	httpx.WriteJSON(w, http.StatusOK, body)
}

func writePayError(w http.ResponseWriter, code, message string) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"returnCode": code, "returnMessage": message})
}

func writeNotification(w http.ResponseWriter, token string, remaining int) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"notificationToken": token,
		"expiresIn":         31536000,
		"remainingCount":    remaining,
		"sessionId":         "line_service_session_mockport",
	})
}

func payInfoResponse(code, message, transactionID string, orderID, amount, currency, status any) linePayResponse {
	return linePayResponse{
		ReturnCode:    code,
		ReturnMessage: message,
		Info: linePayInfo{
			TransactionID: transactionID,
			OrderID:       orderID,
			Amount:        amount,
			Currency:      currency,
			Status:        status,
		},
	}
}
