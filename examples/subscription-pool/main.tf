terraform {
  required_providers {
    azurecnp = {
      source = "hashicorp.com/edu/azurecnp"
    }
  }
}

provider "azurecnp" {}

resource "azurecnp_subscription_pool" "example" {
  pool_management_group_name = "Crossnative"
  target_management_group_name = "cn-hosting"
  subscription_id = "ba18a543-8bb1-4862-a59e-f9eaacbac04f"
  new_subscription_name = "josto-gittem"
}
