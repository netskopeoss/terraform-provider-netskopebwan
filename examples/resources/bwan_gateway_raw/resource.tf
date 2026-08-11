# A gateway is the site-level object. The API only exposes its configuration as
# an opaque JSON document, so the resource has to be asked for explicitly.
#
# Enabling this opts out of compatibility: bwan_gateway_raw is NOT covered by the
# provider's backward-compatibility guarantees, its schema WILL change without a
# major release, and it WILL be removed once a typed bwan_gateway ships.
provider "bwan" {
  enable_raw_gateway = true
}

resource "bwan_gateway_raw" "branch" {
  name  = "branch-01"
  model = "iot-3000"

  # Terraform cannot validate or diff a single field of this document: a change
  # to any part of it is a change to the whole attribute.
  device_config_raw = jsonencode({
    dcfg_schemaver = 1
    dcfg_name      = "branch-01"
  })
}
