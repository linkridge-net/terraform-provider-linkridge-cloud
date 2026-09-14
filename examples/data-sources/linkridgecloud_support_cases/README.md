# linkridgecloud_support_cases Data Source

```hcl
data "linkridgecloud_support_cases" "blocked_import_review" {
  account_id         = "acct_local_qr_demo"
  account_service_id = "asvc_local_qr_demo"
  category           = "import_review"
  status             = "blocked"
}
```

Reads safe support handoff metadata for an account service. This data source
does not create external support tickets, send customer notifications, escalate
incidents, publish customer timelines, or change service status.
