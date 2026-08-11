# This object's whole configuration is a JSON document the API declares no
# shape for, so the resource is off until enable_raw_client_template is set on the
# provider. That opts out of the provider's compatibility guarantees.
resource "netskopebwan_client_template_raw" "example" {
  name      = "example"
  policy_id = "6501f0c2e4b0a1b2c3d4e5f6"
}
