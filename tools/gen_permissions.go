/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package main

import (
	"fmt"
	"strings"
)

func main() {
	queries := "RentalPackage, Currency, WebhookLog, StoreVehicleType, Webhook, OAuthApplication, AdminDashboard, SEOSetting, MarketSetting, JobTimeVariance, JobRequestAcceptanceReport, Job, ProviderLogReport, UserWalletReport, StoreReview, CancelledReport, ProviderPaymentReport, StorePaymentReport, AdminReport, WineDeliveryLabel, GroceryDeliveryLabel, FoodDeliveryLabel, GeneralLabel, AirportSurcharge, LocationWiseFare, DeliveryCharge, DeclineAlert, HelpCategory, HelpDetail, FAQCategory, FAQ, NewsletterSubscriber, EnterpriseAccount, BusinessTripReason, RideProfileType, VisitLocation, VehicleModel, VehicleMake, SMSTemplate, EmailTemplate, GeoFenceRestrictedArea, GeoFenceLocation, DeliveryChargesUtility, OrderStatusUtility, Order, StoreItemType, StoreItem, StoreItemCategory, DeliveryVehicleType, Store, AdvertisementBanner, View, CancelReason, PackageType, Page, User, Review, Coupon, ServiceType, ServiceSubCategory, Service, RequiredDocument, ServiceProvider, ServiceCompany, IAMGroup, MarketStatistics, AppInstallation, Wallet, ServiceVehicleType, ServiceProviderVehicle"
	mutations := "AppInstallation, ServiceProvider, User, UserLocation, ProviderLocation, ServiceCompany, ServiceProvider, Service, ServiceSubCategory, ServiceType, Coupon, CancelReason, Review, PushNotification, Page, PackageType, ServiceProviderVehicle, ServiceVehicleType, BookingFareEstimate, AdvertisementBanner, Store, AppVersion, DeliveryVehicleType, StoreItemCategory, StoreItem, StoreItemType, Order, OrderStatusUtility, DeliveryChargesUtility, GeoFenceLocation, GeoFenceRestrictedArea, EmailTemplate, SMSTemplate, VehicleMake, VehicleModel, VisitLocation, EnterpriseAccount, RideProfileType, BusinessTripReason, Country, State, City, File, DeliveryCharge, LocationWiseFare, AirportSurcharge, GeneralLabel, FoodDeliveryLabel, GroceryDeliveryLabel, WineDeliveryLabel, FAQ, FAQCategory, HelpDetail, HelpCategory, MarketSettings, OAuthApplication, AccessToken, Webhook, Currency, RentalPackage, StoreVehicleType, RequiredDocument, Document"
	queryPermissions := []string{"Read", "List"}
	mutationPermissions := []string{"Create", "Update", "Delete", "Upload"}
	q := strings.Split(strings.Replace(queries, " ", "", -1), ",")
	m := strings.Split(strings.Replace(mutations, " ", "", -1), ",")
	for _, service := range q {
		for _, permission := range queryPermissions {
			fmt.Println("\"" + service + ":" + permission + "\"" + ",")
		}
	}
	for _, service := range m {
		for _, permission := range mutationPermissions {
			if permission == "Upload" && service != " File" {
				break
			}
			fmt.Println("\"" + service + ":" + permission + "\"" + ",")
		}
	}
}
