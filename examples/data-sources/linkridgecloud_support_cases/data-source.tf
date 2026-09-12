data "linkridgecloud_support_cases" "blocked_import_review" {
  account_id         = "acct_local_qr_demo"
  account_service_id = "asvc_local_qr_demo"
  category           = "import_review"
  status             = "blocked"
}
