terraform {
  required_providers {
    bwan = {
      source = "netskope/bwan"
    }
  }
}

# endpoint and token default to $BWAN_ENDPOINT and $BWAN_TOKEN, which is the
# usual way to keep the token out of the configuration.
provider "bwan" {
  endpoint = "https://foo.api.infiot.net"
}
