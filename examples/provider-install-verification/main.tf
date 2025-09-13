terraform {
  required_providers {
    hashicups = {
      source = "hashicorp.com/edu/azurecnp"
    }
  }
}

provider "hashicups" {}

data "hashicups_coffees" "example" {}
