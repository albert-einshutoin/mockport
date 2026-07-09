package stripe

const (
	scenarioPaymentSuccess = "payment_success"
	scenarioPaymentFailed  = "payment_failed"
	scenarioAuthError      = "auth_error"
	scenarioRateLimited    = "rate_limited"
	scenarioTimeout        = "timeout"
)

func normalizeScenario(scenario string) string {
	if scenario == "" {
		return scenarioPaymentSuccess
	}
	return scenario
}
