// Command terraform-provider-bwan serves the Netskope Borderless WAN Terraform
// provider.
//
// Its resources and data sources are generated from the BWAN v2 OpenAPI
// document; see README.md for how the pipeline fits together.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"infiot.com/infiot/mgmt/tf-provider/internal/provider"
)

// address is how Terraform refers to this provider in a configuration's
// required_providers block.
const address = "registry.terraform.io/netskope/bwan"

// version is stamped at build time; it only ends up in diagnostics and the
// User-Agent header.
var version = "dev"

func main() {
	debug := flag.Bool("debug", false, "run the provider attached to a debugger and print the reattach configuration")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: address,
		Debug:   *debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
