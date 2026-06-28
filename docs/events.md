# Event System & Webhook Integration

PicFast provides a unified event system that emits domain events (image upload, processing, deletion, moderation review, user registration) to registered webhooks.

## Event Envelope

All events follow a common JSON envelope:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "image.uploaded",
  "version": "2026-06-01",
  "occurred_at": "2026-06-28T12:34:56.789Z",
  "idempotency_key": "image.uploaded:42",
  "actor": { "kind": "user", "id": "7", "email": "alice@example.com" },
  "subject": { "kind": "image", "id": "42", "key": "aBcDeFgH" },
  "data": { }
}
```

Consumers should treat the `idempotency_key` as the unique identifier for deduplication purposes. The same `idempotency_key` may be re-emitted; receivers must be idempotent.

## Event Types

| Type | Version | Description |
|------|---------|-------------|
| `image.uploaded` | `2026-06-01` | An image was successfully uploaded |
| `image.processed` | `2026-06-01` | Image processing pipeline completed (compression, thumbnail, moderation) |
| `image.deleted` | `2026-06-01` | An image was deleted |
| `moderation.reviewed` | `2026-06-01` | Moderation status changed to approved or rejected |
| `user.registered` | `2026-06-01` | A new user registered |
| `webhook.test` | `2026-06-01` | Test event sent from the webhook management console |

## Versioning Policy

Events follow an additive-only versioning scheme:

- **`version`** uses an ISO date string (e.g. `2026-06-01`)
- Fields are only **added** to data payloads; existing fields are never removed or renamed
- Breaking changes use a **new `type`** or a **major version bump** in the version string
- Consumers should ignore unknown fields to stay forward-compatible

## Webhook Delivery

### HTTP Request

```
POST {webhook.url}
Content-Type: application/json
User-Agent: PicFast-Webhook/1.0
X-PicFast-Event-Id: {envelope.id}
X-PicFast-Event-Type: {envelope.type}
X-PicFast-Delivery-Id: {delivery.id}
X-PicFast-Timestamp: {unix_seconds}
X-PicFast-Signature: sha256={hex_hmac}
```

Body: Full event envelope JSON.

### Signature Verification

```
signed_payload = "{timestamp}.{raw_body}"
signature = HMAC-SHA256(webhook_secret, signed_payload)
header = "sha256=" + hex(signature)
```

Receivers should:
1. Reject requests where `|now - timestamp| > 300s` (replay protection)
2. Use constant-time comparison for the signature
3. Use `idempotency_key` for deduplication

### Retry Policy

Failed deliveries are retried with exponential backoff:

| Attempt | Delay |
|---------|-------|
| 1 | Immediate |
| 2 | 1 minute |
| 3 | 5 minutes |
| 4 | 30 minutes |
| 5 | 2 hours |
| 6 | 12 hours |

After 6 failed attempts, the delivery is marked as dead.

### Example: Node.js Verification

```js
const crypto = require('crypto');

function verifySignature(secret, timestamp, rawBody, expectedSignature) {
  const payload = `${timestamp}.${rawBody}`;
  const hmac = crypto.createHmac('sha256', secret);
  hmac.update(payload);
  const computed = `sha256=${hmac.digest('hex')}`;
  return crypto.timingSafeEqual(Buffer.from(computed), Buffer.from(expectedSignature));
}

// Usage in Express/HTTP handler
app.post('/webhook', (req, res) => {
  const sig = req.headers['x-picfast-signature'];
  const ts = req.headers['x-picfast-timestamp'];
  const body = JSON.stringify(req.body);

  if (Math.abs(Date.now()/1000 - parseInt(ts)) > 300) {
    return res.status(400).send('Timestamp too old');
  }

  if (!verifySignature(process.env.WEBHOOK_SECRET, ts, body, sig)) {
    return res.status(401).send('Invalid signature');
  }

  const event = req.body;
  console.log(`Received event: ${event.type}`);
  res.status(200).send('OK');
});
```

### Example: n8n Webhook Node

1. Add a "Webhook" node in n8n, set HTTP Method to `POST`
2. Copy the production webhook URL
3. In PicFast, create a webhook with the n8n URL and select event types
4. In n8n, add a "Webhook" node → set Response Mode to "Last Node"
5. Connect to processing nodes

### Example: curl

```bash
# Create a webhook
curl -X POST http://localhost:8080/api/v1/webhooks \
  -H "Authorization: Bearer $PICFAST_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"n8n","url":"https://n8n.example.com/webhook/picfast","events":["image.uploaded"]}'

# List webhooks
curl http://localhost:8080/api/v1/webhooks \
  -H "Authorization: Bearer $PICFAST_TOKEN"

# Get delivery logs
curl http://localhost:8080/api/v1/webhooks/1/deliveries \
  -H "Authorization: Bearer $PICFAST_TOKEN"

# Replay a failed delivery
curl -X POST http://localhost:8080/api/v1/webhook-deliveries/1/replay \
  -H "Authorization: Bearer $PICFAST_TOKEN"
```
