package model

type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
)

type SubscriptionPlan string

const (
	SubscriptionPlanFree    SubscriptionPlan = "free"
	SubscriptionPlanPremium SubscriptionPlan = "premium"
)
