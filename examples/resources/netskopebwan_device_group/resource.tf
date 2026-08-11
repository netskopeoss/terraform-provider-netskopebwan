resource "netskopebwan_device_group" "example" {
  address      = "<address>"
  address_type = "ipv4"
  group_name   = "example"
  segment_id   = 1
}
