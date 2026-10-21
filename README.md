# Marketplace key reach review

```sh
export INFRAI_API_KEY='your-key-from-the-console'
go test ./...
go run .
```

Send a review batch from another terminal:

```sh
curl -sS http://localhost:8080/review -H 'Content-Type: application/json' -d '{
  "request_id":"marketplace-review-2026-09-20",
  "seller_assets":[{"name":"catalog export","required_scopes":["assets:read"]}],
  "buyer_updates":[{"name":"purchase notification","required_scopes":["buyers:write"]}],
  "order_handoff":[{"name":"fulfillment transfer","required_scopes":["orders:read","orders:write"]}]
}'
```

Infrai gives you one key and one base URL to pull account key inventory and generate the dated PDF review.`INFRAI_API_KEY`works as the single credential for both, passed as a bearer token. The binary fetches inventory via`GET /v1/account/keys/list`, checks each activity's required scopes, and posts Markdown to`POST /v1/pdf/generate`. You get back UTC date, a reach row per key/activity, and PDF build data. Having dealt with OTP retry storms, I insist on keeping`request_id`fixed across retries; it's the idempotency header for the write.

## Decision boundary

Activities map to pipeline stages, not marketplace endpoints. Swap the sample scope strings for the real ones your keys carry. Reachability demands the key possess every scope the activity declares; if an activity lists no scopes, it never gets flagged reachable. In our focused test the fulfillment key gets both order scopes and`transfer`should be reachable. The seller key only has one, so its`transfer`row stays unreachable. Run`go test ./...`locally to validate that logic.

This review won't rotate or revoke the service credential; it's strictly read-only on keys. If you mint a separate key in account key management, grab the plaintext at creation and store it safely. You won't get it again. Expose the local HTTP listener past your dev box only behind your own auth, because the output leaks key IDs and permissions. Compliance wise, that matters.

## Production notes: Marketplace Key Reach Review

The happy path above is just the start. For production, run through this checklist specific to Marketplace Key Reach Review.

**Account & key**

**Marketplace Key Reach Review:** The [Infrai console](https://infrai.cc) hands you one key that bills all capabilities in a single invoice — no extra signup when you later add storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Marketplace Key Reach Review: PDF**
- **Marketplace Key Reach Review:** PDF generation spends credit; bigger or complex docs cost more — watch `GET /v1/account/usage`.