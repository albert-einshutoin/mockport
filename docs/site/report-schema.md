# Report Schema

Mockport exposes a **runtime report** at `GET /_mockport/report`. The JSON shape is stable for CI and local tooling. It describes what a Mockport process recorded during a run. It is **not** a provider API response and does not mirror Stripe, OpenAI, GitHub, or other upstream schemas.

The same snapshot is available through the CLI:

```bash
mockport report --format json
```

For report usage and request-history limits, see [Reports](reports.md) and [Reporting](../reporting.md).

## Top-level fields

| Field | Type | Description |
| --- | --- | --- |
| `mode` | string | Active Mockport mode, such as `ai-safe` or `strict`. |
| `safety` | object | AI-safe summary counts. See [Safety summary](#safety-summary). |
| `adapters` | array | Enabled adapters with base paths, scenarios, maturity, and capabilities. |
| `requests` | array | Bounded request metadata recorded during the run. Bodies and secret headers are not stored by default. |
| `request_history` | object | Buffer summary for `requests`: `limit`, `retained`, `evicted`, and `truncated`. |
| `safety_warnings` | array | Detailed safety findings with `field`, `category`, and `message`. |
| `scenario_coverage` | array | Supported scenarios per adapter. |
| `behavior_matrix` | array | Supported endpoints per adapter, with maturity, method, path, supported scenarios, and optional notes. |
| `compatibility` | array | Measured compatibility level, score, provider version, SDK/client evidence, and known gaps. Omitted when empty. |
| `state_coverage` | array | Stateful resource coverage, idempotency, and reset support per adapter. Omitted when empty. |
| `unsupported_endpoints` | array | Requests whose recorded `reason` marks unsupported behavior. Derived from bounded `requests`. |

## Safety summary

The `safety` object is a compact trust summary for CI gates and local checks.

| Field | Type | Description |
| --- | --- | --- |
| `mode` | string | Same mode value as top-level `mode`. |
| `safe` | boolean | `true` when there are no safety warnings. |
| `real_looking_secrets` | number | Count of warnings with category `real_looking_secret`. |
| `external_urls` | number | Count of warnings with category `external_url`. |
| `public_env_safe` | boolean | `true` when no real-looking secret or external provider URL warning exists. Deployment and unsupported-config warnings do not change this field. |

### `real_looking_secrets`

Increments once per `safety_warnings` entry whose `category` is `real_looking_secret`. Mockport flags config values that look like live provider credentials, including common API key prefixes for supported providers. Use `mockport_`-prefixed fake values in config to stay clear of this count.

### `external_urls`

Increments once per `safety_warnings` entry whose `category` is `external_url`. Mockport flags live provider base URLs in config. Local URLs such as `http://localhost:43101` are expected and do not increment this count.

### `public_env_safe`

`true` when the run has no `real_looking_secret` or `external_url` warnings. Treat `false` as a signal that the configuration may contain values that are unsafe to commit, even though request bodies and secret headers are not stored by default.

Other warning categories, such as `public_bind` and `unsupported_config`, still appear in `safety_warnings` and make `safe` false. They describe deployment or configuration behavior rather than whether environment values are safe to commit, so they do not change `public_env_safe`.

## Adapter and request objects

### `adapters[]`

| Field | Type | Description |
| --- | --- | --- |
| `name` | string | Adapter id, such as `stripe`. |
| `base_path` | string | Mounted HTTP prefix, such as `/stripe`. |
| `enabled` | boolean | Whether the adapter is enabled in config. |
| `scenario` | string | Active scenario name when configured. |
| `maturity` | string | Adapter maturity label. See [Support matrix](support-matrix.md). |
| `capabilities` | array | Optional capability labels advertised by the adapter. |

### `requests[]`

| Field | Type | Description |
| --- | --- | --- |
| `id` | number | Monotonic request id within the run. |
| `timestamp` | string | UTC timestamp in RFC3339 format. |
| `method` | string | HTTP method. |
| `path` | string | Request path. |
| `status` | number | HTTP response status recorded for the request. |
| `adapter` | string | Adapter that handled the request, when known. |
| `scenario` | string | Active scenario when the request was recorded. |
| `reason` | string | Classification such as `unsupported_endpoint` when applicable. |

### `unsupported_endpoints[]`

Each entry mirrors a request that has a non-empty `reason`, typically unsupported route or method attempts. The list uses the same bounded history as `requests`.

| Field | Type | Description |
| --- | --- | --- |
| `method` | string | HTTP method. |
| `path` | string | Request path. |
| `status` | number | HTTP status returned. |
| `reason` | string | Unsupported classification recorded for the request. |

## Coverage and compatibility

### `scenario_coverage[]`

| Field | Type | Description |
| --- | --- | --- |
| `adapter` | string | Adapter id. |
| `scenarios` | array | Scenario entries with `name` and `supported`. |

### `behavior_matrix[]`

| Field | Type | Description |
| --- | --- | --- |
| `adapter` | string | Adapter id. |
| `maturity` | string | Adapter maturity at report time. |
| `method` | string | HTTP method. |
| `path` | string | Route pattern or path. |
| `supported_scenarios` | array | Scenario names that exercise the route. |
| `notes` | string | Optional human-readable note. |

### `compatibility[]`

| Field | Type | Description |
| --- | --- | --- |
| `adapter` | string | Adapter id. |
| `level` | string | Measured compatibility level. |
| `score` | number | Compatibility score from the manifest. |
| `endpoint_coverage` | number | Optional endpoint coverage percentage. |
| `scenario_coverage` | number | Optional scenario coverage percentage. |
| `sdk_coverage` | number | Optional SDK coverage percentage. |
| `state_coverage` | number | Optional state coverage percentage. |
| `error_coverage` | number | Optional error-behavior coverage percentage. |
| `promotion_eligible` | boolean | Whether the adapter meets promotion rules. |
| `provider_version` | string | Provider API version recorded in the manifest. |
| `sdk_versions` | array | Optional SDK versions used as evidence. |
| `client_evidence` | array | Optional client contract evidence labels. |
| `contract_evidence` | object | Optional fixtures, SDK contracts, and known gaps. |
| `unsupported_endpoints` | array | Optional manifest endpoint ids still unsupported. |

`contract_evidence` objects use:

| Field | Type | Description |
| --- | --- | --- |
| `fixtures` | array | Fixture paths supporting the claim. |
| `sdk_contracts` | array | SDK contract identifiers. |
| `known_gaps` | array | Documented known-gap references. |

### `state_coverage[]`

| Field | Type | Description |
| --- | --- | --- |
| `adapter` | string | Adapter id. |
| `stateful_resources` | array | Stateful resource names tracked by the adapter. |
| `idempotency` | boolean | Whether idempotency-key replay is supported. |
| `reset` | boolean | Whether adapter state can be reset for tests. |

## Example snapshot

The snippet below is sanitized for public docs. Values use `mockport_` fake secrets and `localhost` URLs only.

```json
{
  "mode": "ai-safe",
  "safety": {
    "mode": "ai-safe",
    "safe": true,
    "real_looking_secrets": 0,
    "external_urls": 0,
    "public_env_safe": true
  },
  "adapters": [
    {
      "name": "stripe",
      "base_path": "/stripe",
      "enabled": true,
      "scenario": "payment_success",
      "maturity": "workflow-compatible",
      "capabilities": ["checkout_sessions", "payment_intents", "webhooks"]
    }
  ],
  "requests": [
    {
      "id": 1,
      "timestamp": "2026-07-08T00:00:00Z",
      "method": "POST",
      "path": "/stripe/v1/checkout/sessions",
      "status": 200,
      "adapter": "stripe",
      "scenario": "payment_success"
    }
  ],
  "request_history": {
    "limit": 500,
    "retained": 1,
    "evicted": 0,
    "truncated": false
  },
  "safety_warnings": [],
  "scenario_coverage": [
    {
      "adapter": "stripe",
      "scenarios": [
        {"name": "payment_success", "supported": true},
        {"name": "payment_failed", "supported": true}
      ]
    }
  ],
  "behavior_matrix": [
    {
      "adapter": "stripe",
      "maturity": "workflow-compatible",
      "method": "POST",
      "path": "/stripe/v1/checkout/sessions",
      "supported_scenarios": ["payment_success", "payment_failed"]
    }
  ],
  "compatibility": [
    {
      "adapter": "stripe",
      "level": "workflow",
      "score": 100,
      "endpoint_coverage": 100,
      "scenario_coverage": 100,
      "sdk_coverage": 100,
      "state_coverage": 100,
      "error_coverage": 100,
      "promotion_eligible": true,
      "provider_version": "2025-10-29.clover",
      "sdk_versions": ["stripe@22.3.1"],
      "client_evidence": ["stripe-node checkout session"]
    }
  ],
  "state_coverage": [
    {
      "adapter": "stripe",
      "stateful_resources": [
        "checkout_session",
        "payment_intent",
        "customer",
        "product",
        "price",
        "subscription",
        "invoice",
        "refund"
      ],
      "idempotency": true,
      "reset": true
    }
  ],
  "unsupported_endpoints": []
}
```

Real runs may omit empty `compatibility` or `state_coverage` arrays. Request history defaults to 500 retained entries unless `MOCKPORT_REQUEST_HISTORY` overrides it with a positive integer.
