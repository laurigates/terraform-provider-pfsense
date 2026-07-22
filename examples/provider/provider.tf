# Configuring the pfSense provider.
#
# The provider talks to the pfSense-pkg-RESTAPI (REST API v2) package on the
# firewall, authenticating with an API key sent as the X-API-Key header.

terraform {
  required_providers {
    pfsense = {
      source  = "laurigates/pfsense"
      version = "~> 0.1"
    }
  }
}

# Preferred: leave `host` and `api_key` unset and supply them through the
# PFSENSE_HOST / PFSENSE_API_KEY environment variables, so no credential
# ever lands in the configuration or in state.
provider "pfsense" {
  insecure = true # pfSense ships a self-signed certificate by default
}

# Alternatively, set them explicitly — always from a variable, never a literal.
variable "pfsense_api_key" {
  description = "pfSense REST API key (X-API-Key)."
  type        = string
  sensitive   = true
}

provider "pfsense" {
  alias = "explicit"

  host     = "https://192.168.1.1"
  api_key  = var.pfsense_api_key
  insecure = true
}
