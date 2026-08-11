# Generated from the provider's schema: the arguments this object requires, with
# placeholders to fill in. Remove these two lines to maintain the example by hand.
resource "netskopebwan_vpnpeer" "example" {
  description = "Managed by Terraform"
  ikev2 = {
    dh_group          = 1
    dpd_timeout       = 1
    encryption        = "any"
    hash              = "any"
    ike_sa_lifetime   = 1
    ipsec_sa_lifetime = 1
  }
  ip_address = "<ip_address>"
  location = {
    lat = 1
    lng = 1
  }
  name = "example"
}
