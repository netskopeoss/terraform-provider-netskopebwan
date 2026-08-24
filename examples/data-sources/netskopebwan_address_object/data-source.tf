# Generated from the provider's schema: the arguments this object requires, with
# placeholders to fill in. Remove these two lines to maintain the example by hand.
# A single object is addressed either by its id...
data "netskopebwan_address_object" "by_id" {
  group_id = "6501f0c2e4b0a1b2c3d4e5f6"
  id       = "6501f0c2e4b0a1b2c3d4e5f6"
}

# ...or by a filter, which has to match exactly one object.
data "netskopebwan_address_object" "by_filter" {
  filter   = "name eq \"example\""
  group_id = "6501f0c2e4b0a1b2c3d4e5f6"
}
