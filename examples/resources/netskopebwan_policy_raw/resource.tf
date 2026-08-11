# Generated from the provider's schema: the arguments this object requires, with
# placeholders to fill in. Remove these two lines to maintain the example by hand.
# This object's whole configuration is a JSON document the API declares no
# shape for, so the resource is off until enable_raw_policy is set on the
# provider. That opts out of the provider's compatibility guarantees.
resource "netskopebwan_policy_raw" "example" {
  destinations_config_raw       = jsonencode({})
  disabled                      = false
  name                          = "example"
  policy_config_raw             = jsonencode({})
  productivity_score_config_raw = jsonencode({})
  type                          = "client"
}
