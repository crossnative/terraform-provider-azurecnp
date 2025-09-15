terraform {
  required_providers {
    azurecnp = {
      source = "hashicorp.com/edu/azurecnp"
    }
  }
}

provider "azurecnp" {}

resource "azurecnp_subscription_pool_lease" "example" {
  pool_management_group_name = "Crossnative"
  pool_subscription_prefix = "Azure_Subscription_Crossnative_Pool_"
  target_management_group_name = "db5d4f4b-72c0-4f83-a37c-cd44305348ce"
  target_subscription_name = "josto-test"
}
