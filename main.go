// Package main provides the Terraform provider binary entry point.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/charpand/terraform-provider-mailboxorg/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Run the docs generation tool.
//go:generate go tool tfplugindocs generate --provider-name mailboxorg --rendered-website-dir docs --website-source-dir templates/docs

var (
	// these will be set by the goreleaser configuration to appropriate values
	// for the compiled binary.
	version = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/charpand/mailboxorg",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err.Error())
	}
}
