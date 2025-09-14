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
  target_management_group_name = "cn-hosting"
  target_subscription_name = "josto-gottem"
}
