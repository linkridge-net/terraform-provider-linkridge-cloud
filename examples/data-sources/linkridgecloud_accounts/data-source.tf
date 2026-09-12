provider "linkridgecloud" {
  base_url = "https://dev.cloud.linkridge.net"
}

data "linkridgecloud_accounts" "draft" {
  status = "draft"
}
