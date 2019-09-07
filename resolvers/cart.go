/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"encoding/base64"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/audit_log"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

var (
	ErrProductNotFound          = errors.New("product not found")
	ErrProductVariationNotFound = errors.New("product variation not found")
)

//func (r *mutationResolver) AddCart(ctx context.Context, input models.AddCartInput) (*models.Cart, error) {
//	cart := &models.Cart{}
//	_ = copier.Copy(&cart, &input)
//	user, err := auth.ForContext(ctx)
//	if err != nil {
//		return nil, err
//	}
//	cart.CreatedBy = user.ID
//	cart, err = models.CreateCart(*cart)
//	if err != nil {
//		return nil, err
//	}
//	return cart, nil
//}
//
//func (r *mutationResolver) UpdateCart(ctx context.Context, input models.UpdateCartInput) (*models.Cart, error) {
//	cart := &models.Cart{}
//	cart, err := models.GetCartByID(input.ID.Hex())
//	if err != nil {
//		return nil, err
//	}
//	_ = copier.Copy(&cart, &input)
//
//	user, err := auth.ForContext(ctx)
//	if err != nil {
//		return nil, err
//	}
//	cart.CreatedBy = user.ID
//	cart, err = models.UpdateCart(cart)
//	if err != nil {
//		return nil, err
//	}
//	return cart, nil
//}

type cartResolver struct{ *Resolver }

type cartItemResolver struct{ *Resolver }

func (r cartItemResolver) Product(ctx context.Context, obj *models.CartItem) (*models.Product, error) {
	//TODO: fetch variation id if exists.
	return models.GetProductByID(obj.ProductID.Hex()), nil
}

func (r cartResolver) User(ctx context.Context, obj *models.Cart) (*models.User, error) {
	return models.GetUserByID(obj.UserID.Hex()), nil
}

func (r cartResolver) StoreID(ctx context.Context, obj *models.Cart) (primitive.ObjectID, error) {
	storeID, err := primitive.ObjectIDFromHex(obj.StoreID)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return storeID, err
}

func (r cartResolver) CartItemsQuantity(ctx context.Context, obj *models.Cart) (int, error) {
	return len(obj.Items), nil
}

func (r cartResolver) CartTotal(ctx context.Context, obj *models.Cart) (float64, error) {
	var cartTotal float64
	for _, item := range obj.Items {
		if item.VariationID != nil {
			productVariation, err := models.GetProductVariationByID(*item.VariationID)
			if err != nil {
				return 0, ErrProductVariationNotFound
			}
			cartTotal = cartTotal + productVariation.SalePrice
		}
		if item.VariationID == nil && !item.ProductID.IsZero() {
			product := models.GetProductByID(item.ProductID.Hex())
			if product.ID.IsZero() {
				return 0, ErrProductNotFound
			}
			cartTotal = cartTotal + product.SalePrice
		}
	}

	return cartTotal, nil
}

func (r *mutationResolver) DeleteCart(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteCartByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "cart", nil, nil, ctx)
	return &res, err
}

func (r *queryResolver) Carts(ctx context.Context, cartID primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.CartConnection, error) {
	var items []*models.Cart
	var edges []*models.CartEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetCarts(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.CartEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	cartList := &models.CartConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return cartList, nil
}

func (r *queryResolver) Cart(ctx context.Context, id primitive.ObjectID) (*models.Cart, error) {
	cart := models.GetCartByID(id.Hex())
	return cart, nil
}

func containsObjectID(s []primitive.ObjectID, e primitive.ObjectID) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (r *mutationResolver) AddProductToCart(ctx context.Context, productID primitive.ObjectID, variationID *primitive.ObjectID, quantity int) (*models.Cart, error) {
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//TODO : identifies user by store id
	cart := &models.Cart{}
	cart, err = models.GetCartByFilter(bson.D{{"userID", user.ID}})
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	if cart == nil || cart.ID.IsZero() {
		cart, err = models.CreateCart(&models.Cart{UserID: user.ID, CreatedBy: user.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()})
		if err != nil {
			return nil, err
		}
		//Update audit log
		go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), cart.ID.Hex(), "cart", cart, nil, ctx)
	}
	cartItems := cart.Items
	addCartItem := &models.CartItem{
		ProductID:   productID,
		VariationID: variationID,
		Quantity:    quantity,
		Type:        models.CartItemTypeProduct,
	}
	if len(cartItems) > 0 {
		for i, item := range cartItems {
			if item.VariationID != nil && addCartItem.VariationID != nil {
				if item.VariationID.Hex() == addCartItem.VariationID.Hex() {
					if item.ProductID == addCartItem.ProductID {
						product := models.GetProductByID(addCartItem.ProductID.Hex())
						if product.ID.IsZero() {
							return cart, ErrProductNotFound
						}
						if !containsObjectID(product.Variations, *variationID) {
							return cart, ErrProductVariationNotFound
						}
						if addCartItem.Quantity == 0 {
							//remove(cartItems,addCartItem.VariationID)
						}
						cartItems[i].Quantity = addCartItem.Quantity
						break
					}
				}
			}
			if item.VariationID == nil && addCartItem.VariationID == nil && !item.ProductID.IsZero() && item.ProductID == addCartItem.ProductID {
				if addCartItem.Quantity == 0 {
					//remove(cartItems,addCartItem.VariationID)
				}
				cartItems[i].Quantity = addCartItem.Quantity
				break
			}
		}
	} else {
		cartItems = append(cartItems, *addCartItem)
	}
	cart.Items = cartItems
	updatedCart, err := models.UpdateCart(cart)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), updatedCart.ID.Hex(), "cart", updatedCart, nil, ctx)
	return updatedCart, nil
}
