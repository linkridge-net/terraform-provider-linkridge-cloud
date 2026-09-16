# linkridgecloud_provisioning_runs Data Source

```hcl
data "linkridgecloud_provisioning_runs" "rehearsal" {
  account_service_id = "asvc_local_qr_demo"
  status             = "rehearsal_only"
}
```

Reads safe provisioning rehearsal evidence. This data source does not approve
provisioning, perform external writes, or enable customer-visible service
access.
