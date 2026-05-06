package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/rnikoopour/terraform-provider-pihole/internal/provider"
)

var version string = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run the provider with support for debuggers")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/rnikoopour/pihole",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
