# A cloud account holds the credentials for one cloud, and the API declares one
# shape of credentials per cloud: aws, azure, gcp, netskope or device_security.
# Terraform has no type for "one of these", so they are sibling blocks and exactly
# one of them is set — `terraform validate` says so if none or several are.
#
# `cloud_provider` is the one renamed argument in the provider: the API calls the
# field `provider`, which Terraform reserves, so it is exposed under this name and
# translated on the wire.
resource "netskopebwan_cloud_account" "aws" {
  name           = "production"
  cloud_provider = "aws"

  config = {
    aws = {
      key_id            = var.aws_key_id
      secret_access_key = var.aws_secret_access_key
    }
  }
}
