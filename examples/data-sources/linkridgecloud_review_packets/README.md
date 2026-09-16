# linkridgecloud_review_packets Data Source

Reads normalized internal approval-gate evidence from `/v1/review-packets`.
This data source is read-only: it does not approve work, execute imports, issue
secrets, export billing usage, create support tickets, notify customers, or
enable customer-visible effects.

```hcl
data "linkridgecloud_review_packets" "service_token" {
  account_id  = "acct_local_qr_demo"
  packet_type = "service_token"
  status      = "blocked_pending_matthew_approval"
}
```
