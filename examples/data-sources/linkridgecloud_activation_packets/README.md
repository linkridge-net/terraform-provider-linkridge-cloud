# linkridgecloud_activation_packets Data Source

```hcl
data "linkridgecloud_activation_packets" "planned" {
  account_id         = "acct_local_qr_demo"
  account_service_id = "asvc_local_qr_demo"
  status             = "blocked_pending_matthew_approval"
}
```

Reads safe activation review metadata for an account service. This data source
does not approve activation, create billing/customer access, issue secrets, send
invites, or run external provisioning.
