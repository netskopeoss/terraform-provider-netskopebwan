# A single object is addressed either by its id...
data "netskopebwan_ca_certificate" "by_id" {
  id = "6501f0c2e4b0a1b2c3d4e5f6"
}

# ...or by a filter, which has to match exactly one object.
data "netskopebwan_ca_certificate" "by_filter" {
  filter = "name eq \"example\""
}
