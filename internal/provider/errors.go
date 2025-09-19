package provider

import "fmt"

type NoSubscriptionsFoundError struct {
	SubscriptionId string
}

func (n NoSubscriptionsFoundError) Error() string {
	return fmt.Sprintf("No subscriptions found for ID '%s'", n.SubscriptionId)
}

func NewNoSubscriptionsFoundError(subscriptionId string) NoSubscriptionsFoundError {
	return NoSubscriptionsFoundError{
		SubscriptionId: subscriptionId,
	}
}
