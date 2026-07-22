//go:build generate

// Package tools pins the documentation-generation toolchain in its own module
// so tfplugindocs never enters the provider's dependency graph.
package tools

// -provider-name must match the registry address suffix in main.go's ServeOpts
// (registry.terraform.io/laurigates/pfsense), since it determines the
// pfsense_* type-name prefix stripped from generated page names.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-dir .. -provider-name pfsense

import _ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
