package resolvers

// Resolver ...
type Resolver struct{}

//BookingResolver
type BookingResolver interface{}

//Cart
func (r *Resolver) Cart() CartResolver {
	return &cartResolver{r}
}

//Job
func (r *Resolver) Job() JobResolver {
	return &jobResolver{r}
}

//CartItem
func (r *Resolver) CartItem() CartItemResolver {
	return &cartItemResolver{r}
}

//Product
func (r *Resolver) Product() ProductResolver {
	return &productResolver{r}
}

//ProductReview
func (r *Resolver) ProductReview() ProductReviewResolver {
	return &productReviewResolver{r}
}

//ProductVariation
//func (r *Resolver) ProductVariation() ProductVariationResolver {
//	return &productVariationResolver{r}
//}

//Service
func (r *Resolver) Service() ServiceResolver {
	return &serviceResolver{r}
}

//Store
func (r *Resolver) Store() StoreResolver {
	return &storeResolver{r}
}

// DeclineAlertForProvider resolver
func (r *Resolver) DeclineAlertForProvider() DeclineAlertForProviderResolver {
	return &declineAlertForProviderResolver{r}
}

// FAQ resolver
func (r *Resolver) FAQ() FAQResolver {
	return &fAQResolver{r}
}

// Installation resolver
func (r *Resolver) Installation() InstallationResolver {
	return &installationResolver{r}
}

// Order resolver
func (r *Resolver) Order() OrderResolver {
	return &orderResolver{r}
}

// PackageType resolver
func (r *Resolver) PackageType() PackageTypeResolver {
	return &packageTypeResolver{r}
}

// Query resolver
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

// Review resolver
func (r *Resolver) Review() ReviewResolver {
	return &reviewResolver{r}
}

// ServiceProvider resolver
func (r *Resolver) ServiceProvider() ServiceProviderResolver {
	return &serviceProviderResolver{r}
}

// Subscription resolver
func (r *Resolver) Subscription() SubscriptionResolver {
	return &subscriptionResolver{r}
}

// User resolver
func (r *Resolver) User() UserResolver {
	return &userResolver{r}
}

// WalletTransaction resolver
func (r *Resolver) WalletTransaction() WalletTransactionResolver {
	return &walletTransactionResolver{r}
}

// Webhook resolver
func (r *Resolver) Webhook() WebhookResolver {
	return &webhookResolver{r}
}

// Withdrawal resolver
func (r *Resolver) Withdrawal() WithdrawalResolver {
	return &withdrawalResolver{r}
}

// ServiceCompany resolver
func (r *Resolver) ServiceCompany() ServiceCompanyResolver {
	return &serviceCompanyResolver{r}
}

// Country resolver
func (r *Resolver) Country() CountryResolver {
	return &countryResolver{r}
}

// Document resolver
func (r *Resolver) Document() DocumentResolver {
	return &documentResolver{r}
}

// Mutation resolver
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// mutationResolver represents mutation resolver.
type mutationResolver struct{ *Resolver }

// queryResolver represents query resolver.
type queryResolver struct{ *Resolver }
