# Webhook Notifications

JVS can send HTTP webhook notifications for important events like snapshot creation, restore operations, garbage collection, and verification.

## Configuration

Webhooks are configured in `.jvs/config.yaml`:

```yaml
webhooks:
  enabled: true
  max_retries: 3
  retry_delay: 5s
  async_queue_size: 100

  hooks:
    - url: https://example.com/webhook
      secret: your-hmac-secret-key
      events:
        - snapshot.created
        - snapshot.deleted
        - restore.complete
      timeout: 10s
      enabled: true
```

## Event Types

| Event | Description |
|-------|-------------|
| `snapshot.created` | Fired when a snapshot is created |
| `snapshot.deleted` | Fired when a snapshot is deleted |
| `restore.start` | Fired when a restore operation starts |
| `restore.complete` | Fired when a restore completes successfully |
| `restore.failed` | Fired when a restore fails |
| `gc.start` | Fired when garbage collection starts |
| `gc.complete` | Fired when garbage collection completes |
| `verify.start` | Fired when verification starts |
| `verify.complete` | Fired when verification completes successfully |
| `verify.failed` | Fired when verification fails for a snapshot |

Use `*` as the event name to receive all events.

## Event Payload

All webhook payloads have the following structure:

```json
{
  "event": "snapshot.created",
  "timestamp": "2024-02-23T10:00:00Z",
  "repo_id": "550e8400-e29b-41d4-a716-446655440000",
  "repo_root": "/path/to/repo",
  "snapshot_id": "abc123...",
  "note": "Initial snapshot",
  "tags": ["v1.0", "release"],
  "error": "",
  "metadata": {}
}
```

### Event-Specific Fields

#### snapshot.created / snapshot.deleted
- `snapshot_id`: The ID of the snapshot
- `note`: The snapshot note
- `tags`: Array of tags

#### restore.start / restore.complete / restore.failed
- `snapshot_id`: The target snapshot ID
- `error`: Error message (for `restore.failed`)

#### gc.complete
- `metadata.freed_bytes`: Bytes freed
- `metadata.snapshots_deleted`: Number of snapshots deleted

#### verify.complete / verify.failed
- `snapshot_id`: The verified snapshot ID (for `verify.failed`)
- `metadata.snapshots_verified`: Total snapshots verified (for `verify.complete`)
- `error`: Error message (for `verify.failed`)

## Security

### HMAC Signatures

When you configure a `secret`, JVS signs each webhook payload using HMAC-SHA256. The signature is sent in the `X-JVS-Signature` header:

```
X-JVS-Signature: sha256=abcdef1234567890...
```

To verify the signature on your server:

```python
import hmac
import hashlib

def verify_signature(payload, signature, secret):
    expected = hmac.new(
        secret.encode(),
        payload,
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(f"sha256={expected}", signature)
```

## Examples

### Slack Integration

Send notifications to Slack using an incoming webhook:

```yaml
webhooks:
  enabled: true
  hooks:
    - url: https://hooks.slack.com/services/YOUR/WEBHOOK/URL
      events:
        - snapshot.created
        - restore.complete
```

### Discord Integration

For Discord, format the payload differently using a middleware service.

### Custom Webhook Server

A simple Go server to receive webhooks:

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
)

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    var payload map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, err.Error(), 400)
        return
    }

    event := payload["event"].(string)
    log.Printf("Received event: %s", event)

    // Verify signature if secret is configured
    signature := r.Header.Get("X-JVS-Signature")
    if signature != "" {
        // Verify signature
    }

    w.WriteHeader(http.StatusOK)
}

func main() {
    http.HandleFunc("/webhook", webhookHandler)
    log.Println("Server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Retries and Failures

JVS will retry failed webhook deliveries up to `max_retries` times with exponential backoff. Failed deliveries are logged to the JVS output channel.

Webhooks are sent asynchronously by default, so JVS operations won't be blocked by slow webhook endpoints.

## Troubleshooting

### Webhooks Not Firing

1. Check that webhooks are enabled in `config.yaml`
2. Verify the webhook URL is accessible
3. Check JVS logs for error messages

### Signature Verification Failing

1. Ensure the secret matches exactly
2. Verify you're using the raw request body for signature calculation
3. Check that you're comparing with the `sha256=` prefix

### Missing Events

1. Verify the event names match exactly (see Event Types table)
2. Use `*` to receive all events for testing
3. Check that the specific hook is `enabled: true`
