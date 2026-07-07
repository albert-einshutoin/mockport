package report

import (
	"bytes"
	"fmt"
	"strings"
)

func RenderText(snapshot Snapshot) string {
	var out bytes.Buffer
	fmt.Fprintln(&out, "Mockport Report")
	fmt.Fprintln(&out)
	fmt.Fprintf(&out, "Mode: %s\n", snapshot.Mode)
	fmt.Fprintf(&out, "Safety: safe=%v real-looking-secrets=%d external-urls=%d\n", snapshot.Safety.Safe, snapshot.Safety.RealLookingSecrets, snapshot.Safety.ExternalURLs)
	fmt.Fprintf(&out, "Public env safe-to-commit: %v\n", snapshot.Safety.PublicEnvSafe)
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "Adapters:")
	for _, adapter := range snapshot.Adapters {
		if adapter.Enabled {
			fmt.Fprintf(&out, "- %s enabled at %s", adapter.Name, adapter.BasePath)
			if adapter.Maturity != "" {
				fmt.Fprintf(&out, " maturity=%s", adapter.Maturity)
			}
			fmt.Fprintln(&out)
		}
	}
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "Requests:")
	for _, request := range snapshot.Requests {
		fmt.Fprintf(&out, "- #%d %s %s -> %d", request.ID, request.Method, request.Path, request.Status)
		if request.Reason != "" {
			fmt.Fprintf(&out, " reason=%s", request.Reason)
		}
		fmt.Fprintln(&out)
	}
	if snapshot.RequestHistory.Truncated {
		fmt.Fprintf(&out, "Request history: truncated (limit=%d retained=%d evicted=%d)\n", snapshot.RequestHistory.Limit, snapshot.RequestHistory.Retained, snapshot.RequestHistory.Evicted)
	}
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "Scenario coverage:")
	for _, coverage := range snapshot.ScenarioCoverage {
		for _, scenario := range coverage.Scenarios {
			fmt.Fprintf(&out, "- %s %s supported=%v\n", coverage.Adapter, scenario.Name, scenario.Supported)
		}
	}
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "Behavior matrix:")
	for _, behavior := range snapshot.BehaviorMatrix {
		fmt.Fprintf(&out, "- %s %s %s maturity=%s\n", behavior.Adapter, behavior.Method, behavior.Path, behavior.Maturity)
	}
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "Compatibility:")
	for _, compatibility := range snapshot.Compatibility {
		fmt.Fprintf(&out, "- %s level=%s score=%d provider=%s unsupported=%d\n", compatibility.Adapter, compatibility.Level, compatibility.Score, compatibility.ProviderVersion, len(compatibility.UnsupportedEndpoints))
	}
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "State coverage:")
	for _, coverage := range snapshot.StateCoverage {
		fmt.Fprintf(&out, "- %s resources=%s idempotency=%v reset=%v\n", coverage.Adapter, strings.Join(coverage.StatefulResources, ","), coverage.Idempotency, coverage.Reset)
	}
	fmt.Fprintln(&out)
	fmt.Fprintf(&out, "Unsupported endpoints: %d\n", len(snapshot.UnsupportedEndpoints))
	fmt.Fprintf(&out, "Safety warnings: %d\n", len(snapshot.SafetyWarnings))
	return out.String()
}
