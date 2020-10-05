package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/kinvolk/terraform-provider-tinkerbell/tinkerbell"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: tinkerbell.Provider,
	})
}
