# linkridgecloud_qr_mutation_rehearsal_requirements Data Source

```hcl
data "linkridgecloud_qr_mutation_rehearsal_requirements" "menu_archive" {
  workspace_id        = "qrw_local_growth_demo"
  mutation_request_id = "qrm_local_growth_demo_menu_archive"
}
```

Reads safe internal/dev QR mutation rehearsal requirements. This data source
does not approve work, execute QR mutations, write tenant QR records, enable
hosted redirects, record billing usage, or create customer-visible effects.
