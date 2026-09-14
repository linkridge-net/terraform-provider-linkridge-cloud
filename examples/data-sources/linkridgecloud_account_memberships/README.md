# linkridgecloud_account_memberships Data Source

Reads account-scoped role assignment evidence from the LinkRidge Cloud control
plane. This data source does not create users, send invites, grant external
access, or change roles.

```hcl
data "linkridgecloud_account_memberships" "owners" {
  account_id = "acct_local_qr_demo"
  role       = "owner"
  status     = "planned"
}
```
