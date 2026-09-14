# linkridgecloud_audit_events Data Source

Reads append-only control-plane audit evidence from `/v1/audit-events`.
This data source does not replay events, approve work, issue secrets, execute
imports, export billing usage, notify customers, or enable customer-visible
effects.

```hcl
data "linkridgecloud_audit_events" "service_token_prepare" {
  account_id       = "acct_local_qr_demo"
  action           = "service_token.prepare"
  source_packet_id = "local-qr-starter-demo"
}
```
