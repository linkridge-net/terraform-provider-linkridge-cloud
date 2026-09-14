// Copyright (c) LinkRidge
// SPDX-License-Identifier: MPL-2.0

// Package main is the entry point for the LinkRidge Cloud Terraform provider.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/linkridge-net/terraform-provider-linkridge-cloud/internal/provider"
)

// version is set by release builds. Local builds use "dev".
var version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with debugger support")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/linkridge-net/linkridge-cloud",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err.Error())
	}
}
