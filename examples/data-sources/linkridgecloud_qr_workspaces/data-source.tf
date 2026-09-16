provider "linkridgecloud" {
  base_url = "https://dev.cloud.linkridge.net"
}

data "linkridgecloud_qr_workspaces" "planned" {
  account_id = "acct_local_qr_demo"
  status     = "planned"
}
