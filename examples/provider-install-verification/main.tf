terraform {
  required_providers {
    azurecnp = {
      source = "hashicorp.com/edu/azurecnp"
    }
  }
}

provider "azurecnp" {}

data "azurecnp_coffees" "example" {}
