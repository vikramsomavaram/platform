/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package models

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/cache"
	"github.com/tribehq/platform/lib/database"
	"github.com/tribehq/platform/utils/webhooks"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// ProductType is the type of a product.
type ProductType string

// List of values that ProductType can take.
const (
	ProductTypeGood    ProductType = "good"
	ProductTypeService ProductType = "service"

	ProductTypeSimple   ProductType = "simple"
	ProductTypeGrouped  ProductType = "grouped"
	ProductTypeExternal ProductType = "external"
	ProductTypeVariable ProductType = "variable"
)

// Product represents a product.
type Product struct {
	ID                primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt         time.Time            `json:"createdAt" bson:"createdAt"`
	DeletedAt         *time.Time           `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt         time.Time            `json:"updatedAt" bson:"updatedAt"`
	CreatedBy         primitive.ObjectID   `json:"createdBy" bson:"createdBy"`
	Name              string               `json:"name" bson:"name"`
	MenuItem          string               `json:"item" bson:"item"`
	Slug              string               `json:"slug" bson:"slug"`
	Permalink         string               `json:"permalink" bson:"permalink"`
	Type              ProductType          `json:"type" bson:"type"`
	Status            string               `json:"status" bson:"status"`
	IsFeatured        bool                 `json:"featured" bson:"featured"`
	ItemTagName       string               `json:"itemTagName" bson:"itemTagName"`
	CatalogVisibility string               `json:"catalogVisibility" bson:"catalogVisibility"`
	Description       string               `json:"description" bson:"description"`
	ShortDescription  string               `json:"shortDescription" bson:"shortDescription"`
	Sku               string               `json:"sku" bson:"sku"`
	Price             float64              `json:"price" bson:"price"`
	RegularPrice      float64              `json:"regularPrice" bson:"regularPrice"`
	ServiceType       StoreCategory        `json:"serviceType" bson:"serviceType"`
	SalePrice         float64              `json:"salePrice" bson:"salePrice"`
	DateOnSaleFrom    time.Time            `json:"dateOnSaleFrom" bson:"dateOnSaleFrom"`
	DateOnSaleTo      time.Time            `json:"dateOnSaleTo" bson:"dateOnSaleTo"`
	PriceHTML         string               `json:"priceHtml" bson:"priceHtml"`
	OnSale            bool                 `json:"onSale" bson:"onSale"`
	Purchasable       bool                 `json:"purchasable" bson:"purchasable"`
	TotalSales        int                  `json:"totalSales" bson:"totalSales"`
	Store             string               `json:"store" bson:"store"`
	Virtual           bool                 `json:"virtual" bson:"virtual"`
	DisplayOrder      int                  `json:"displayOrder" bson:"displayOrder"`
	Downloadable      bool                 `json:"downloadable" bson:"downloadable"`
	Downloads         []ProductDownload    `json:"downloads" bson:"downloads"`
	DownloadLimit     int                  `json:"downloadLimit" bson:"downloadLimit"`
	DownloadExpiry    int                  `json:"downloadExpiry" bson:"downloadExpiry"`
	ExternalURL       string               `json:"externalUrl" bson:"externalUrl"`
	ButtonText        string               `json:"buttonText" bson:"buttonText"`
	TaxStatus         string               `json:"taxStatus" bson:"taxStatus"`
	TaxClass          string               `json:"taxClass" bson:"taxClass"`
	ManageStock       bool                 `json:"manageStock" bson:"manageStock"`
	StockQuantity     int                  `json:"stockQuantity" bson:"stockQuantity"`
	StockStatus       string               `json:"stockStatus" bson:"stockStatus"`
	BackOrders        string               `json:"backOrders" bson:"backOrders"`
	BackOrdersAllowed bool                 `json:"backOrdersAllowed" bson:"backOrdersAllowed"`
	BackOrdered       bool                 `json:"backOrdered" bson:"backOrdered"`
	SoldIndividually  bool                 `json:"soldIndividually" bson:"soldIndividually"`
	Weight            float64              `json:"weight" bson:"weight"`
	Dimensions        ProductDimensions    `json:"dimensions" bson:"dimensions"`
	ShippingRequired  bool                 `json:"shippingRequired" bson:"shippingRequired"`
	ShippingTaxable   bool                 `json:"shippingTaxable" bson:"shippingTaxable"`
	ShippingClass     string               `json:"shippingClass" bson:"shippingClass"`
	ShippingClassID   string               `json:"shippingClassId" bson:"shippingClassId"`
	ReviewsAllowed    bool                 `json:"reviewsAllowed" bson:"reviewsAllowed"`
	AverageRating     string               `json:"averageRating" bson:"averageRating"`
	RatingCount       int                  `json:"ratingCount" bson:"ratingCount"`
	RelatedIds        []string             `json:"relatedIds" bson:"relatedIds"`
	UpsellIds         []string             `json:"upsellIds" bson:"upsellIds"`
	CrossSellIds      []string             `json:"crossSellIds" bson:"crossSellIds"`
	ParentID          string               `json:"parentId" bson:"parentId"`
	PurchaseNote      string               `json:"purchaseNote" bson:"purchaseNote"`
	Categories        []ProductCategory    `json:"categories" bson:"categories"`
	Tags              []ProductTag         `json:"tags" bson:"tags"`
	Images            []ProductImage       `json:"images" bson:"images"`
	Attributes        []ProductAttribute   `json:"attributes" bson:"attributes"`
	DefaultAttributes []ProductAttribute   `json:"defaultAttributes" bson:"defaultAttributes"`
	Variations        []primitive.ObjectID `json:"variations" bson:"variations"`
	GroupedProducts   []primitive.ObjectID `json:"groupedProducts" bson:"groupedProducts"`
	MenuOrder         int                  `json:"menuOrder" bson:"menuOrder"`
	MetaData          []ProductMetadata    `json:"metaData" bson:"metaData"`
	IsActive          bool                 `json:"isActive" bson:"isActive"`
}

//ProductDimensions represents the dimensions of a product.
type ProductDimensions struct {
	Length float64 `json:"length" bson:"length"`
	Width  float64 `json:"width" bson:"width"`
	Height float64 `json:"height" bson:"height"`
}

// CreateProduct creates new products.
func CreateProduct(product Product) (*Product, error) {
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	product.ID = primitive.NewObjectID()
	db := database.MongoDB
	installationCollection := db.Collection(ProductsCollection)
	ctx := context.Background()
	_, err := installationCollection.InsertOne(ctx, &product)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("product.created", &product)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(product.ID.Hex(), product, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &product, nil
}

// GetProductByID gives requested product by id.
func GetProductByID(ID string) *Product {
	db := database.MongoDB
	product := &Product{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(product)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return product
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(ProductsCollection).FindOne(ctx, filter).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return product
		}
		log.Errorln(err)
		return product
	}
	//set cache item
	err = cacheClient.Set(ID, product, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return product
}

// GetProducts gives a list of products.
func GetProducts(filter bson.D, limit int, after *string, before *string, first *int, last *int) (products []*Product, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ProductsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ProductsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		product := &Product{}
		err = cur.Decode(&product)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		products = append(products, product)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return products, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateProduct updates the product.
func UpdateProduct(p *Product) (*Product, error) {
	product := p
	product.UpdatedAt = time.Now()
	filter := bson.D{{"_id", product.ID}}
	db := database.MongoDB
	productsCollection := db.Collection(ProductsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := productsCollection.FindOneAndReplace(context.Background(), filter, product, findRepOpts).Decode(&product)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("product.updated", &product)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(product.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return product, nil
}

// DeleteProductByID deletes product by id.
func DeleteProductByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	productsCollection := db.Collection(ProductsCollection)
	res, err := productsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("product.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (product *Product) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, product); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (product *Product) MarshalBinary() ([]byte, error) {
	return json.Marshal(product)
}

//ProductMetadata represents product metadata.
type ProductMetadata struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Key       string             `json:"key" bson:"key"`
	Value     string             `json:"value" bson:"value"`
}

// CreateProductMetadata creates new productMetadata.
func CreateProductMetadata(productMetadata ProductMetadata) (*ProductMetadata, error) {
	productMetadata.CreatedAt = time.Now()
	productMetadata.UpdatedAt = time.Now()
	productMetadata.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(ProductMetadataCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &productMetadata)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("productMetadata.created", &productMetadata)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(productMetadata.ID.Hex(), productMetadata, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &productMetadata, nil
}

// GetProductMetadataByID gives requested productMetadata by id.
func GetProductMetadataByID(ID string) *ProductMetadata {
	db := database.MongoDB
	productMetadata := &ProductMetadata{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(productMetadata)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return productMetadata
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(ProductMetadataCollection).FindOne(ctx, filter).Decode(&productMetadata)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil
		}
		log.Errorln(err)
		return nil
	}
	//set cache item
	err = cacheClient.Set(ID, productMetadata, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return productMetadata
}

// GetProductMetadatas gives a list of product metadata.
func GetProductMetadatas(filter bson.D, limit int, after *string, before *string, first *int, last *int) (productMetadatas []*ProductMetadata, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ProductMetadataCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ProductMetadataCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		productMetadata := &ProductMetadata{}
		err = cur.Decode(&productMetadata)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		productMetadatas = append(productMetadatas, productMetadata)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return productMetadatas, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateProductMetadata updates product metadata.
func UpdateProductMetadata(c *ProductMetadata) *ProductMetadata {
	productMetadata := c
	productMetadata.UpdatedAt = time.Now()
	filter := bson.D{{"_id", productMetadata.ID}}
	db := database.MongoDB
	productMetadataCollection := db.Collection(ProductMetadataCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := productMetadataCollection.FindOneAndReplace(context.Background(), filter, productMetadata, findRepOpts).Decode(&productMetadata)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("productMetadata.updated", &productMetadata)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(productMetadata.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return productMetadata
}

// DeleteProductMetadataByID deletes productMetadata by id.
func DeleteProductMetadataByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	productMetadataCollection := db.Collection(ProductMetadataCollection)
	res, err := productMetadataCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("productMetadata.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (productMetadata *ProductMetadata) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, productMetadata); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (productMetadata *ProductMetadata) MarshalBinary() ([]byte, error) {
	return json.Marshal(productMetadata)
}

//ProductDownload represents downloaded products.
type ProductDownload struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Name      string             `json:"name" bson:"name"`
	File      string             `json:"file" bson:"file"`
}

// CreateProductDownload creates new product download.
func CreateProductDownload(product ProductDownload) (*ProductDownload, error) {
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	product.ID = primitive.NewObjectID()
	db := database.MongoDB
	installationCollection := db.Collection(ProductDownloadCollection)
	ctx := context.Background()
	_, err := installationCollection.InsertOne(ctx, &product)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("product_download.created", &product)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(product.ID.Hex(), product, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &product, nil
}

// GetProductDownloadByID gives requested product download by id.
func GetProductDownloadByID(ID string) (*ProductDownload, error) {
	db := database.MongoDB
	product := &ProductDownload{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(product)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	filter := bson.D{{"_id", ID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(ProductDownloadCollection).FindOne(ctx, filter).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, product, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return product, nil
}

// GetProductDownloads gives a list of product downloads.
func GetProductDownloads(filter bson.D, limit int, after *string, before *string, first *int, last *int) (prodDownloads []*ProductDownload, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ProductDownloadCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ProductDownloadCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		prodDownload := &ProductDownload{}
		err = cur.Decode(&prodDownload)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		prodDownloads = append(prodDownloads, prodDownload)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return prodDownloads, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateProductDownload updates the product download.
func UpdateProductDownload(product *ProductDownload) *ProductDownload {
	product.UpdatedAt = time.Now()
	filter := bson.D{{"_id", product.ID}}
	db := database.MongoDB
	productsCollection := db.Collection(ProductDownloadCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := productsCollection.FindOneAndReplace(context.Background(), filter, product, findRepOpts).Decode(&product)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("product_download.updated", &product)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(product.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return product
}

// DeleteProductDownloadByID deletes product download by id.
func DeleteProductDownloadByID(ID string) (bool, error) {
	db := database.MongoDB
	filter := bson.D{{"_id", ID}}
	productsCollection := db.Collection(ProductDownloadCollection)
	res, err := productsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("product_download.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (product *ProductDownload) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, product); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (product *ProductDownload) MarshalBinary() ([]byte, error) {
	return json.Marshal(product)
}

//ProductImages represents images of the product.
type ProductImage struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Src       string             `json:"src" bson:"src"`
	Name      string             `json:"name" bson:"name"`
	Alt       string             `json:"alt" bson:"alt"`
}

// CreateProductImage creates new product images.
func CreateProductImage(product ProductImage) (*ProductImage, error) {
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	product.ID = primitive.NewObjectID()
	db := database.MongoDB
	installationCollection := db.Collection(productImageCollection)
	ctx := context.Background()
	_, err := installationCollection.InsertOne(ctx, &product)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("product_image.created", &product)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(product.ID.Hex(), product, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &product, nil
}

// GetProductImageByID gives requested product image by id.
func GetProductImageByID(ID string) *ProductImage {
	db := database.MongoDB
	productImage := &ProductImage{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(productImage)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return productImage
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(productImageCollection).FindOne(ctx, filter).Decode(&productImage)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		log.Errorln(err)
		return nil
	}
	//set cache item
	err = cacheClient.Set(ID, productImage, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return productImage
}

// GetProductImages gives a list of product images.
func GetProductImages(filter bson.D, limit int, after *string, before *string, first *int, last *int) (prodImages []*ProductImage, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(productImageCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(productImageCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		prodImage := &ProductImage{}
		err = cur.Decode(&prodImage)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		prodImages = append(prodImages, prodImage)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return prodImages, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateProductImage updates the product images.
func UpdateProductImage(product *ProductImage) *ProductImage {
	product.UpdatedAt = time.Now()
	filter := bson.D{{"_id", product.ID}}
	db := database.MongoDB
	productsCollection := db.Collection(productImageCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := productsCollection.FindOneAndReplace(context.Background(), filter, product, findRepOpts).Decode(&product)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("product_image.updated", &product)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(product.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return product
}

// DeleteProductImageByID deletes product image by id.
func DeleteProductImageByID(ID string) (bool, error) {
	db := database.MongoDB
	filter := bson.D{{"_id", ID}}
	productsCollection := db.Collection(productImageCollection)
	res, err := productsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("product_image.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (product *ProductImage) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, product); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (product *ProductImage) MarshalBinary() ([]byte, error) {
	return json.Marshal(product)
}
