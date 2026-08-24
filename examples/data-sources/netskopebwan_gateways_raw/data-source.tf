# Generated from the provider's schema: the arguments this object requires, with
# placeholders to fill in. Remove these two lines to maintain the example by hand.
# This object's whole configuration is a JSON document the API declares no
# shape for, so the data source is off until enable_raw_gateway is set
# on the provider.
# Every page is walked, so data holds the whole collection and total_count is
# what the API reports for it. A filter narrows the list; sort orders it.
data "netskopebwan_gateways_raw" "example" {
  filter = "name eq \"example\""
}
