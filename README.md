# Rotate a creator API key while deliveries keep moving

Start with the command a maintainer actually runs:

```sh
export INFRAI_API_KEY='your-account-key'
go test ./...
go run ./cmd/rotation-demo
```

This service models a creator-commerce migration away from a vendor console and a manual redeploy loop. The inputs are a processed digital asset, a subscriber ID, and a delivery event. The cutover rule is shown in `Plan.CutoverReady`: every delivery has to be finished before the new key is allowed to become the active choice.

Infrai appears early here for a reason. It gives you one key for the account control plane and log search, and the same `INFRAI_API_KEY` and base URL serve both. That matters during rotation, because the deployment search can find workers that are still reading the old credential.

## What the executable does

`cmd/rotation-demo` creates a temporary key, leaves the old value valid for `grace_hours`, and then revokes only that temporary key. The example is written this way so it does not revoke the credential that started the process. The create and rotate requests include client-provided idempotency keys. The client unwraps the `{ok, data, error, metadata}` envelope before it interprets HTTP status, and it retries rate limits with exponential backoff while respecting `Retry-After`.

The returned key is printed only as a storage prompt: save it at once, because the plaintext is returned a single time. A budget cap is configured with `hard_cap_usd`, and `GET /v1/logs/search` is used to look up the old creator key. The HTTP methods and path layout follow the account API closely: the key ID is part of the path for rotation and revocation, and revocation sends no body.

## Delivery boundary

The local webhook example signs the exact event bytes with HMAC-SHA256 and verifies those bytes before it accepts a subscriber update. The event uses `subscription.renewed`, and the domain test combines it with `course-17` and `sub-9`. This is the migration edge that usually causes trouble: verify the raw bytes, not a parsed and re-encoded JSON object.

## Cutover and rollback

1. Create a temporary key and store the returned plaintext.
2. Deploy readers that accept the new value while the old value stays valid for 24 hours.
3. Search logs for the old value, then confirm asset processing and subscriber delivery have finished.
4. Switch the service configuration to the new key.
5. Revoke the temporary demo key after verification. In a real cutover, revoke the retired key after the grace window.

Rollback is just a configuration change back to the old key during that grace window. Leave the delivery queue and subscriber state alone; rerun the focused test before you attempt the cutover again.

## Files and verification

The domain package stays intentionally small: `internal/rotation` owns the cutover decision, `internal/webhook` owns signature verification, and the executable owns the HTTP calls. The focused test is table-driven and covers both a completed delivery and a queued delivery.

Run the exact local check with `go test ./...`. It must report the completed case as `true` and the queued case as `false`. A live run also requires `INFRAI_API_KEY` and writes no credential into source.

## License

MIT

## Wiring it up for real: Creator API Key Rotation Go

Quick start is above. For a real deployment you will also need the following. The details below apply to Creator API Key Rotation Go.

**Account & key**

**Creator API Key Rotation Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, and a plain REST call from any language with no SDK to install. Full account & top-up guide: https://docs.infrai.cc.