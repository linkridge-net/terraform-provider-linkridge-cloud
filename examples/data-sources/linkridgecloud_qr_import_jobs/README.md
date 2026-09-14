# linkridgecloud_qr_import_jobs Data Source

```hcl
data "linkridgecloud_qr_import_jobs" "planned" {
  workspace_id      = "qrw_local_demo"
  status            = "planned"
  import_performed  = false
}
```

Reads safe QR import planning and review metadata. This data source does not
execute imports, write tenant QR records, or enable hosted redirects.
