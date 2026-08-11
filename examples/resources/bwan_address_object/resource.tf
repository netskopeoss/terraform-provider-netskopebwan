resource "bwan_address_group" "datacenter" {
  name = "datacenter"
}

# An address object lives inside a group. group_id comes from the API path rather
# than the request body, so moving an object to another group replaces it.
resource "bwan_address_object" "dns" {
  group_id = bwan_address_group.datacenter.id

  name    = "primary-dns"
  address = "10.0.0.53"
  type    = "ipv4"
}

# Importing one takes the parent and the object: terraform import
# bwan_address_object.dns <group_id>/<id>
