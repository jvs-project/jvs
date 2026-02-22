package main

import (
	"context"
	"flag"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/jvs-project/jvs/terraform-provider-jvs/internal/provider"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "start plugin in debug mode")

	flag.Parse()

	// Serve the provider
	providerserver.Serve(
		context.Background(),
		provider.New,
		providerserver.WithDebug(debug),
	)
}
