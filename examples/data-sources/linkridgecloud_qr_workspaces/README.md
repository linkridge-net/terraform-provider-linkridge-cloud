Read LinkRidge Cloud QR workspace planning records.

```hcl
data "linkridgecloud_qr_workspaces" "planned" {
  account_id = "acct_local_qr_demo"
  status     = "planned"
}
```
