terraform {
  required_providers {
    netskopebwan = {
      source = "netskopeoss/netskopebwan"
    }
  }
}

# endpoint and token default to $BWAN_ENDPOINT and $BWAN_TOKEN, which is the
# usual way to keep the token out of the configuration.
provider "netskopebwan" {
  endpoint = "https://foo.api.infiot.net"
}
