data "linkridgecloud_account_invites" "draft_editors" {
  account_id             = "acct_local_qr_demo"
  role                   = "editor"
  status                 = "draft"
  invite_delivery_status = "not_sent"
}
