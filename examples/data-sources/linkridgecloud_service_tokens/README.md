# linkridgecloud_service_tokens Data Source

```hcl
data "linkridgecloud_service_tokens" "planned" {
  account_id             = "acct_local_qr_demo"
  status                 = "planned"
  secret_material_issued = false
}
```

This data source reads safe service-token metadata only. It never returns or creates service-token secret material.
