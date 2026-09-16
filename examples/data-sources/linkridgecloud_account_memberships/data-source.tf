data "linkridgecloud_account_memberships" "owners" {
  account_id = "acct_local_qr_demo"
  role       = "owner"
  status     = "planned"
}
