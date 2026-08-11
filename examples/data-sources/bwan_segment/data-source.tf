# A single object is addressed either by its id...
data "bwan_segment" "by_id" {
  id = "6501f0c2e4b0a1b2c3d4e5f6"
}

# ...or by a filter, which has to match exactly one object.
data "bwan_segment" "by_name" {
  filter = "name eq \"corporate\""
}

output "network_id" {
  value = data.bwan_segment.by_name.network_id
}
