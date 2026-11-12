package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/rnikoopour/terraform-provider-pihole/internal/pihole"
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

	// Terraform closes the provider process after every run, so this is the
	// last chance to hand the API sessions back to Pi-hole.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	pihole.CloseAll(ctx)
	cancel()

	if err != nil {
		log.Fatal(err.Error())
	}
}
