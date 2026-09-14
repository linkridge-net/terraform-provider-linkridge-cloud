# linkridgecloud_billing_export_requests Data Source

```hcl
data "linkridgecloud_billing_export_requests" "blocked" {
  account_id         = "acct_local_qr_demo"
  account_service_id = "asvc_local_qr_demo"
  status             = "blocked"
}
```

Reads safe billing export review metadata for an account service. This data
source does not create billing customers, subscriptions, invoices, external
usage records, or metered usage exports.
