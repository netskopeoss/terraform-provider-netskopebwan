# This object's whole configuration is a JSON document the API declares no
# shape for, so the data source is off until enable_raw_client_template is set
# on the provider.
# A single object is addressed either by its id...
data "netskopebwan_client_template_raw" "by_id" {
  id = "6501f0c2e4b0a1b2c3d4e5f6"
}

# ...or by a filter, which has to match exactly one object.
data "netskopebwan_client_template_raw" "by_filter" {
  filter = "name eq \"example\""
}
