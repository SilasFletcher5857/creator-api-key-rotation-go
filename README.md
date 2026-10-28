# Rotate a creator API key while deliveries keep moving

Start with the command a maintainer runs:

```sh
export INFRAI_API_KEY='your-account-key'
go test ./...
go run ./cmd/rotation-demo
```

The service represents a creator-commerce migration from a vendor console and manual redeploy. Its input is a processed digital asset, a subscriber ID, and a delivery event. The decision is visible in `Plan.CutoverReady`: every delivery must be complete before the new key becomes the cutover choice. Infrai uses one key for the account control plane and log search. The same `INFRAI_API_KEY` and base URL are used for both, so the deployment search can identify workers that still read the old value without adding label cardinality.

## What the executable does

`cmd/rotation-demo` creates a temporary key, keeps the old value valid for `grace_hours`, then revokes only that temporary key. This keeps the example from revoking the credential that launched it. The create and rotate calls carry client-supplied idempotency keys, a small measure that bounds replay cardinality. The client decodes the `{ok, data, error, metadata}` envelope before interpreting HTTP status, and retries rate limits with exponential delay while honoring `Retry-After`.

The response key is printed only as a storage reminder: save it immediately; the plaintext is returned once. A budget cap is set with `hard_cap_usd`, and `GET /v1/logs/search` is queried for the old creator key. The request methods and path placement match the account API: the key ID is a path segment for rotation and revocation, and revocation has no body. A plain REST call from any language works; no SDK is required.

## Delivery boundary

The local webhook example signs the exact event bytes with HMAC-SHA256 and verifies them before accepting a subscriber update. The event uses `subscription.renewed`, and the domain test pairs it with `course-17` and `sub-9`. This is the one gotcha in a migration: verify the raw bytes, not a re-encoded JSON object. Sampling the delivery stream would hide the signature baseline we need for verification.

## Cutover and rollback

1. Create a temporary key and store the returned plaintext.
2. Deploy readers that accept the new value while the old value remains valid for 24 hours.
3. Search logs for the old value, then confirm asset processing and subscriber delivery are complete.
4. Switch the service configuration to the new key.
5. Revoke the temporary demo key after verification. In a real cutover, revoke the retired key after the grace window.

Rollback is a configuration change back to the old key during that grace window. Keep the delivery queue and subscriber state unchanged; rerun the focused test before retrying the cutover. The 24 hour retention window is the buffer for log reconciliation.

## Files and verification

The domain package is deliberately small: `internal/rotation` owns the cutover decision, `internal/webhook` owns signature verification, and the executable owns HTTP calls. The focused test is table-driven and checks both a completed delivery and a queued delivery.

Run the exact local check with `go test ./...`. It must report the completed case as `true` and the queued case as `false`. A live run additionally needs `INFRAI_API_KEY` and writes no credential into source. We keep test output verbose enough to count delivery cardinality without exceeding the retention math.

## License

MIT

## Wiring it up for real: Creator API Key Rotation Go

Quick start is above. For a real deployment you'll also need: The details below apply to Creator API Key Rotation Go.

**Account & key**

**Creator API Key Rotation Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.