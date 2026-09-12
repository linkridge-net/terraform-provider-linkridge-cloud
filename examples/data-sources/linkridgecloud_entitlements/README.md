# linkridgecloud_entitlements Data Source

Reads effective account-service entitlement evidence from the LinkRidge Cloud
control-plane API. This data source is read-only: it does not change feature
gates, sync billing, or enable customer-visible service access.

```hcl
data "linkridgecloud_entitlements" "active_qr_codes" {
  account_id         = "acct_local_qr_demo"
  account_service_id = "asvc_local_qr_demo"
  entitlement_key    = "active_qr_codes"
}
```
