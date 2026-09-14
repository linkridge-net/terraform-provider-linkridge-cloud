# linkridgecloud_operator_approvals Data Source

Reads safe operator approval decision evidence from `/v1/operator-approvals`.
This data source does not approve, reject, execute provisioning, issue secrets,
send invites, export billing usage, create support tickets, configure DNS, run
production deploys, or enable customer-visible effects.

```hcl
data "linkridgecloud_operator_approvals" "blocked" {
  account_id        = "acct_local_qr_demo"
  decision_state    = "blocked"
  required_approval = "matthew"
}
```
