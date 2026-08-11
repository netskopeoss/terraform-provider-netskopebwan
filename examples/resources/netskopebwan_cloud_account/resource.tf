# Generated from the provider's schema: the arguments this object requires, with
# placeholders to fill in. Remove these two lines to maintain the example by hand.
resource "netskopebwan_cloud_account" "example" {
  cloud_provider = "aws"
  config = {
    aws = {
      key_id            = "<key_id>"
      secret_access_key = "<secret_access_key>"
    }
  }
  name = "example"
}
