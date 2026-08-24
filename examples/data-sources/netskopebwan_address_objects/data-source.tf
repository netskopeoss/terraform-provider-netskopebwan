# Generated from the provider's schema: the arguments this object requires, with
# placeholders to fill in. Remove these two lines to maintain the example by hand.
# Every page is walked, so data holds the whole collection and total_count is
# what the API reports for it. A filter narrows the list; sort orders it.
data "netskopebwan_address_objects" "example" {
  filter   = "name eq \"example\""
  group_id = "6501f0c2e4b0a1b2c3d4e5f6"
}
