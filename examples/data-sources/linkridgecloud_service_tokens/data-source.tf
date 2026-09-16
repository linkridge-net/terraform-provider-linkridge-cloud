data "linkridgecloud_service_tokens" "planned" {
  account_id             = "acct_local_qr_demo"
  status                 = "planned"
  secret_material_issued = false
}
