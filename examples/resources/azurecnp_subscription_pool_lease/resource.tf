terraform {
  required_providers {
    azurecnp = {
      source = "crossnative/azurecnp"
    }
  }
}

provider "azurecnp" {
  subscription_pool_management_group = "Crossnative"
  subscription_pool_name_prefix = "Azure_Subscription_Crossnative_Pool_"
}

resource "azurecnp_subscription_pool_lease" "example" {
  target_management_group_name = "cn-hosting"
  target_subscription_name = "josto-test"
}
