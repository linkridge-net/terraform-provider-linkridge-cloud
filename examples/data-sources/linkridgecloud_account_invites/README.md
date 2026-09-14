# linkridgecloud_account_invites Data Source

Reads draft account invite evidence from the LinkRidge Cloud control plane. This
data source does not send invites, create external identity access, notify
customers, or approve invite delivery.

```hcl
data "linkridgecloud_account_invites" "draft_editors" {
  account_id             = "acct_local_qr_demo"
  role                   = "editor"
  status                 = "draft"
  invite_delivery_status = "not_sent"
}
```
