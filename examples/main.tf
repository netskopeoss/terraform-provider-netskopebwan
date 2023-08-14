//  Policy Data Source 

data "netskopebwan_policy" "test" {
   name = "aws-ph-policy"
}

// Gateway Resource 
resource "netskopebwan_gateway" "test" {
   name = "001-terraform"
   model = "iXVirtual"
   role = "hub"
   assigned_policy {
      id = data.netskopebwan_policy.test.id 
      name = data.netskopebwan_policy.test.name
   }
}

// Gateway Interface Static IP Config
resource "netskopebwan_gateway_interface" "GE2" {
  gateway_id = resource.netskopebwan_gateway.test.id
  name = "GE2"
  type = "ethernet"
  mode = "routed"
  is_disabled = false
  addresses {
       address = "172.15.1.1"
       address_assignment = "static"
       address_family = "ipv4"
       dns_primary = "8.8.8.8"
       dns_secondary = "8.8.4.4"
       gateway = "172.15.1.254"
       mask = "255.255.255.0"
  }
  do_advertise = true
  enable_nat = false
  mtu = 1400
  zone = "trusted"
}

// Gateway Interface DHCP Config
resource "netskopebwan_gateway_interface" "GE3" {
  gateway_id = resource.netskopebwan_gateway.test.id
  name = "GE3"
  is_disabled = false
}

// BGP Peer
/*
resource "netskopebwan_gateway_bgpconfig" "tgwpeer" {
   gateway_id = resource.netskopebwan_gateway.test.id
   name = "tgwpeer-1"
   neighbor = "169.254.1.1"
   remote_as = 64513
   local_as = 400
}*/

resource "netskopebwan_gateway_bgpconfig" "tgwpeer2" {
   gateway_id = resource.netskopebwan_gateway.test.id
   name = "tgwpeer-2"
   neighbor = "169.254.1.2"
   remote_as = 64512
   local_as = 400
}

// Static Route
resource "netskopebwan_gateway_staticroute" "subnet" {
   gateway_id = resource.netskopebwan_gateway.test.id
   advertise = true
   destination = "54.54.54.100/32"
   device = "GE1"
   install = true
   nhop = "192.168.31.2"
}

// NAT Rule
resource "netskopebwan_gateway_nat" "nat" {
   gateway_id = resource.netskopebwan_gateway.test.id
   name = "test"
   public_ip = "1.1.1.1"
   up_link_if_name = "GE1"
   lan_ip = "192.168.1.10"
   bi_directional = true
}
resource "netskopebwan_gateway_port_forward" "nat" {
   gateway_id = resource.netskopebwan_gateway.test.id
   name = "test"
   public_ip = "1.1.1.1"
   up_link_if_name = "GE1"
   lan_ip = "192.168.1.10"
   bi_directional = true
   public_port = 80
   lan_port = 9000
}

// Gateway Activation Token
resource "netskopebwan_gateway_activate" "token" {
  gateway_id = resource.netskopebwan_gateway.test.id
}

// Policies
resource "netskopebwan_policy" "gwpolicy" {
   name = "gwpolicy1"
   type ="gateway"
}

resource "netskopebwan_policy" "clientpolicy" {
   name = "clientpolicy1"
   type ="client"
}

// Templates
resource "netskopebwan_gateway" "clienttemplate" {
   name = "clienttemplate1"
   description ="example"

   model="Client"
   role="spoke"
   is_template=true

   client_configuration {
      ipv4_pool_ranges {
         pool_start = "10.200.11.1"
         pool_end = "10.200.11.2"
      }
      ipv4_pool_ranges {
         pool_start = "10.200.16.10"
         pool_end = "10.200.16.20"
      }
      ipv4_pool_ranges {
         pool_start = "10.200.20.3"
         pool_end = "10.200.20.10"
      }
   }
}

resource "netskopebwan_gateway" "gwtemplate" {
   name = "gwtemplate1"
   description ="example"

   model="iXVirtual"
   role="spoke"
   is_template=true
}
