data "linkridgecloud_entitlements" "active_qr_codes" {
  account_id         = "acct_local_qr_demo"
  account_service_id = "asvc_local_qr_demo"
  entitlement_key    = "active_qr_codes"
}
