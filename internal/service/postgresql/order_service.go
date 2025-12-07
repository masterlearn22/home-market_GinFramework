package service

import (
	"errors"
	"fmt"
	"log"
	"time"
	entity "home-market/internal/domain"
	mongorepo "home-market/internal/repository/mongodb"
	repo "home-market/internal/repository/postgresql"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


type OrderService struct {
	orderRepo repo.OrderRepository
	shopRepo  repo.ShopRepository 
	logRepo   mongorepo.LogRepository 
}

func NewOrderService(orderRepo repo.OrderRepository, shopRepo repo.ShopRepository, logRepo mongorepo.LogRepository) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		shopRepo: shopRepo,
		logRepo: logRepo,
	}
}

func (s *OrderService) createAndSaveNotification(userID uuid.UUID, title string, message string, notiType string, relatedID uuid.UUID) {
    noti := &entity.Notification{
        ID: primitive.NewObjectID(),
        UserID: userID,
        Title: title,
        Message: message,
        Type: notiType,
        RelatedID: relatedID,
        IsRead: false,
        CreatedAt: time.Now(),
    }
    
    if err := s.logRepo.SaveNotification(noti); err != nil {
        log.Printf("Warning: failed to save notification for user %s: %v", userID.String(), err)
    }
}

// @Summary      Get Marketplace Items
// @Description  Retrieves a list of active items from the marketplace, filtered by keyword, category, and price range.
// @Tags         Marketplace
// @Accept       json
// @Produce      json
// @Param        keyword query string false "Search keyword"
// @Param        category_id query string false "Filter by Category ID (UUID)"
// @Param        min_price query number false "Minimum price filter"
// @Param        max_price query number false "Maximum price filter"
// @Param        limit query integer false "Limit (default 10)"
// @Param        offset query integer false "Offset"
// @Success      200  {array}   entity.Item
// @Failure      500  {object}  map[string]interface{}
// @Router       /market/items [get]
func (s *OrderService) GetMarketplaceItems(filter entity.ItemFilter) ([]entity.Item, error) {
	return s.orderRepo.GetMarketItems(filter)
}

// @Summary      Get Item Detail
// @Description  Retrieves detailed information for a single active item in the marketplace.
// @Tags         Marketplace
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Item ID"
// @Success      200  {object}  entity.Item
// @Failure      404  {object}  map[string]interface{} "Item not found or inactive"
// @Failure      500  {object}  map[string]interface{}
// @Router       /market/items/{id} [get]
func (s *OrderService) GetItemDetail(itemID uuid.UUID) (*entity.Item, error) {
	item, err := s.orderRepo.GetItemForOrder(itemID)
	if err != nil {
		return nil, err
	}
	if item == nil || item.Status != "active" {
		return nil, errors.New("item not found or inactive")
	}
	return item, nil
}

// @Summary      Create New Order
// @Description  Allows a Buyer to create a new order, performing stock validation and decrement within a transaction. Supports only single-shop orders currently.
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        input body entity.CreateOrderInput true "Order details"
// @Success      201  {object}  entity.Order
// @Failure      400  {object}  map[string]interface{} "Validation error (stock, multi-shop)"
// @Failure      403  {object}  map[string]interface{} "Forbidden (not buyer)"
// @Failure      500  {object}  map[string]interface{}
// @Router       /orders [post]
func (s *OrderService) CreateOrder(buyerID uuid.UUID, input entity.CreateOrderInput) (*entity.Order, error) {
	shopItems := make(map[uuid.UUID][]entity.OrderItem)
	
	for _, itemInput := range input.Items {
		item, err := s.orderRepo.GetItemForOrder(itemInput.ItemID)
		if err != nil { return nil, errors.New("database error during item fetch") }
		if item == nil || item.Status != "active" || item.Stock < itemInput.Quantity {
			return nil, errors.New("invalid item, insufficient stock, or item inactive")
		}

		orderItem := entity.OrderItem{
			ItemID: item.ID, Quantity: itemInput.Quantity, Price: item.Price, OrderID: uuid.Nil,
		}
		shopItems[item.ShopID] = append(shopItems[item.ShopID], orderItem)
	}

	if len(shopItems) != 1 {
		return nil, errors.New("multi-shop orders are not supported in a single transaction yet")
	}

	var shopID uuid.UUID
	var itemsForOrder []entity.OrderItem
	var totalPrice float64

	for id, items := range shopItems {
		shopID = id
		itemsForOrder = items
		for _, item := range items {
			totalPrice += item.Price * float64(item.Quantity)
		}
		break
	}
    
    shopOwnerID, err := s.shopRepo.GetShopOwnerID(shopID) 
    if err != nil && err.Error() != "shop not found" {
        log.Printf("Warning: failed to retrieve shop owner ID for notification: %v", err)
    }

	order := &entity.Order{
		ID: uuid.New(), BuyerID: buyerID, ShopID: shopID, TotalPrice: totalPrice, Status: "pending", 
		ShippingAddress: input.ShippingAddress, ShippingCourier: input.ShippingCourier,
	}

	for i := range itemsForOrder {
		itemsForOrder[i].OrderID = order.ID
		itemsForOrder[i].ID = uuid.New()
	}

	if err := s.orderRepo.CreateOrderTransaction(order, itemsForOrder); err != nil {
		return nil, err
	}
    
    if shopOwnerID != uuid.Nil {
        s.createAndSaveNotification(
            shopOwnerID, "Order Baru Masuk",
            fmt.Sprintf("Anda menerima order baru #%s dengan total %.2f.", order.ID.String()[:8], order.TotalPrice),
            "new_order", order.ID,
        )
    }
	return order, nil
}

// @Summary      Update Order Status
// @Description  Allows Seller or Admin to update the status of an order.
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Order ID"
// @Param        input body entity.UpdateOrderStatusInput true "New status value (e.g., paid, processing, cancelled)"
// @Success      200  {object}  entity.Order "Returns updated order"
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{} "Unauthorized"
// @Failure      404  {object}  map[string]interface{} "Order not found"
// @Router       /orders/{id}/status [patch]
func (s *OrderService) UpdateOrderStatus(userID uuid.UUID, role string, orderID uuid.UUID, status string) (*entity.Order, error) {
	if !ValidOrderStatuses[status] {
		return nil, errors.New("invalid status value")
	}
    
    order, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil { return nil, err }
	if order == nil { return nil, errors.New("order not found") }
	shop, _ := s.shopRepo.GetByUserID(userID)
	isOwner := shop != nil && order.ShopID == shop.ID
	isAdmin := role == "admin"
	
	if !isOwner && !isAdmin {
		return nil, errors.New("unauthorized: you are not the shop owner or admin")
	}
	if err := s.orderRepo.UpdateOrderStatus(orderID, status); err != nil { return nil, err }
    order.Status = status 

    s.createAndSaveNotification(
        order.BuyerID, "Status Order Berubah",
        fmt.Sprintf("Status order Anda #%s telah diperbarui menjadi %s.", orderID.String()[:8], status),
        "order_status", order.ID,
    )

	return order, nil
}

// @Summary      Input Shipping Receipt
// @Description  Allows Seller or Admin to input courier and receipt number, automatically setting status to 'shipped'.
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Order ID"
// @Param        input body entity.InputShippingReceiptInput true "Courier and Receipt details"
// @Success      200  {object}  entity.Order "Returns updated order"
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{} "Unauthorized"
// @Failure      404  {object}  map[string]interface{} "Order not found"
// @Router       /orders/{id}/shipping [post]
func (s *OrderService) InputShippingReceipt(userID uuid.UUID, role string, orderID uuid.UUID, input entity.InputShippingReceiptInput) (*entity.Order, error) {
    order, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil { return nil, err }
	if order == nil { return nil, errors.New("order not found") }

	shop, _ := s.shopRepo.GetByUserID(userID)
	isOwner := shop != nil && order.ShopID == shop.ID
	isAdmin := role == "admin"
	
	if !isOwner && !isAdmin { return nil, errors.New("unauthorized") }

	if err := s.orderRepo.UpdateOrderShipment(orderID, input.ShippingCourier, input.ShippingReceipt); err != nil { return nil, err }
    order.ShippingCourier = input.ShippingCourier 
    order.ShippingReceipt = input.ShippingReceipt
    order.Status = "shipped"
    s.createAndSaveNotification(
        order.BuyerID, "Barang Anda Dikirim",
        fmt.Sprintf("Order Anda #%s telah dikirim dengan resi %s.", orderID.String()[:8], input.ShippingReceipt),
        "order_status", order.ID,
    )

	return order, nil
}

// @Summary      Get Order Tracking Details
// @Description  Retrieves order details and associated items for tracking purposes (Buyer or Admin access).
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Order ID"
// @Success      200  {object}  map[string]interface{} "Returns order and order_items"
// @Failure      403  {object}  map[string]interface{} "Unauthorized"
// @Failure      404  {object}  map[string]interface{} "Order not found"
// @Router       /orders/{id}/tracking [get]
func (s *OrderService) GetOrderTracking(userID uuid.UUID, role string, orderID uuid.UUID) (*entity.Order, []entity.OrderItem, error) {
	order, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil { return nil, nil, err }
	if order == nil { return nil, nil, errors.New("order not found") }

	isAdmin := role == "admin"
	isBuyer := order.BuyerID == userID
	if !isBuyer && !isAdmin { return nil, nil, errors.New("unauthorized: access denied") }
	items, err := s.orderRepo.GetOrderItems(orderID)
	if err != nil { return order, nil, err }


	return order, items, nil
}