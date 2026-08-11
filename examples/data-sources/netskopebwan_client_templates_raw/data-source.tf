# This object's whole configuration is a JSON document the API declares no
# shape for, so the data source is off until enable_raw_client_template is set
# on the provider.
# Every page is walked, so data holds the whole collection. A filter narrows
# it; first or after ask for one page instead.
data "netskopebwan_client_templates_raw" "example" {
  filter = "name eq \"example\""
}
