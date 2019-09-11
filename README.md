# Tribe Commerce - Open Source On-Demand Services &amp; E-Commerce Marketplace Platform

[![CodeFactor](https://www.codefactor.io/repository/github/tribehq/platform/badge)](https://www.codefactor.io/repository/github/tribehq/platform)
[![Build Status](https://travis-ci.com/tribehq/platform.svg?branch=master)](https://travis-ci.com/tribehq/platform)
[![codecov](https://codecov.io/gh/tribehq/platform/branch/master/graph/badge.svg)](https://codecov.io/gh/tribehq/platform) 
[![Go Report Card](https://goreportcard.com/badge/github.com/tribehq/platform)](https://goreportcard.com/report/github.com/tribehq/platform)
[![GoDoc](https://godoc.org/github.com/tribehq/platform?status.svg)](https://godoc.org/github.com/tribehq/platform)
[![GitHub release](https://img.shields.io/github/release/tribehq/platform.svg)](https://github.com/tribehq/platform/releases/latest)
[![Join the community on Spectrum](https://withspectrum.github.io/badge/badge.svg)](https://spectrum.chat/tribe)
![Discord](https://img.shields.io/discord/619455063584145409)

Tribe® is an Open-Source, Real-Time, Extensible, On-Demand Commerce Platform built with Golang. You can use tribe to build anything from Uber to Postmates or Amazon like marketplace itself for on-demand services and e-commerce.

### GraphQL Playground : https://dev.graph.tribe.cab
### OAuth2 Server: https://dev.accounts.tribe.cab
### Mobile Apps Repository: https://github.com/tribehq/mobile-apps

## Technology Stack and Requirements
* Golang >= 1.12
* Redis as cache store
* MongoDB >= 4.0 as data store.
* Flutter for cross platform (iOS and Android) mobile applications.
* Headless Commerce framework (Backend APIs/Server) developed using [GQLgen](https://github.com/99designs/gqlgen) framework / library, which supports GraphQL & WebSockets API via GraphQL subscriptions.
* Google Cloud products such as Cloud Pubsub & Cloud Functions are used keeping scalability in mind.
* Frontend & Admin Panel are developed in Vue.JS Framework.

## Features
*   Modern & Open Platform for On-Demand Economy
*   Supports everything that Uber, Lyft could do now ;).
*   Supports fleet tracking, Service Provider on-boarding etc.,
*   Supports Single-Store and Multi-Store / Multi-Vendor / Peer-to-Peer Marketplaces
*   Everything Reactive, Real-Time and Blazing Fast!
*   Headless Commerce framework, which allows different implementations of store-fronts, Admin UIs and client apps. It exposes rich GraphQL and WS APIs.
*   Realtime Webhooks for event driven commerce and integration with other systems / applications.
*   Mobile ordering App for customers to make On-Demand orders (iOS and Android using Flutter & Native technologies)
*   Partner (Driver) Mobile App for deliveries by carriers, drivers or service providers (iOS and Android using Flutter & Native technologies)
*   Customizing Shopping e-commerce Website for customers to make in-browser On-Demand purchases of food, goods or services
*   Merchant Tablet App for Stores/Merchants/Warehouses to manage & track orders, organize deliveries, etc.
*   Admin Website used to manage all platform features and settings in the single Web-based interface
*   Multi-language and culture settings across Platform (i18N)
*   Products Catalogs (global and per Merchant) with Multiple Product Images
*   Inventory/Stock Management and Real-time Order Management/Processing across the Platform
*   Deliveries/Shipping management and processing across Platform (shipping with real-time location tracking for On-Demand orders)
*   Real-Time discounts, promotions and products/services availability updates
*   Customers registration, Guest Checkouts, Wallets , Invitations (optional)
*   Gateway and Payment Processing (currently planned Payment Gateways - Stripe, Braintree Payments, RazorPay, AliPay, Yandex.Checkout)
*   Plugins / Extensions / Custom Fields
*   Firebase Analytics & Notifications
*   Tax Calculations
*   Third-party Shipping providers integrations
*   Users Roles / Permissions across Platform
*   Large products catalogs with products variants, facets and full-text search

### Work In Progress

A word of caution: We are very much under development (work in progress, WIP).
Expect lots of changes and bugs.

## Support
This repository is not suitable for support. Please don't use our issue tracker for support requests, but for core Tribe Platform issues only. Support can take place through the appropriate channels like discord and tribe community forum.

For enterprise installation & support / customizations please mail to support@tribe.cab

## Sales Channels (planned / supported)

* Mobile Apps
* Web
* Facebook Messenger
* Telegram
* Alexa
* Google Home / Actions
* Apple Siri
* WhatsApp
* Anything you can imagine and build for :)


## Contributing to Tribe Commerce

If you have a patch or have stumbled upon an issue with tribe platform, you can contribute this back to the code. Please read our [contributor guidelines](https://github.com/tribehq/platform/blob/master/.github/CONTRIBUTING.md) for more information how you can do this.

## Contributors 👨🏽‍💻
* Balamurali Pandranki
* Vikram Somavaram
* Srikanth Koppuravuri
* Chaitanya Rayampally
* Sanjana Argula
* Ramyasai Sanjita Bhavirisetty

And many more awesome [contributors listed here](https://github.com/tribehq/platform/graphs/contributors) 
