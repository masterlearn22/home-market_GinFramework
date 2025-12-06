package repository

import (
	"database/sql"
	entity "home-market/internal/domain"
    "errors"
	"github.com/google/uuid"
)

type ItemRepository interface {
	CreateItem(item *entity.Item) error
	CreateItemImage(img *entity.ItemImage) error
	GetShopByUserID(userID uuid.UUID) (*entity.Shop, error)
    GetShopOwnerID(shopID uuid.UUID) (uuid.UUID, error)
	IsCategoryOwnedByShop(categoryID, shopID uuid.UUID) (bool, error)
	GetItemByID(id uuid.UUID) (*entity.Item, error)
    UpdateItem(item *entity.Item) error
	CreateOffer(offer *entity.Offer) error
    GetOffersByGiverID(giverID uuid.UUID) ([]entity.Offer, error)
    GetOffersBySellerID(sellerID uuid.UUID) ([]entity.Offer, error)
    GetOfferByID(offerID uuid.UUID) (*entity.Offer, error)
    UpdateOffer(offer *entity.Offer) error
    GetMarketItems(filter entity.ItemFilter) ([]entity.Item, error)
    GetItemForOrder(itemID uuid.UUID) (*entity.Item, error) // Ambil detail item + stok
    CreateOrderTransaction(order *entity.Order, items []entity.OrderItem) error
    GetOrderByID(orderID uuid.UUID) (*entity.Order, error)
    UpdateOrderStatus(orderID uuid.UUID, status string) error
    UpdateOrderShipment(orderID uuid.UUID, courier string, receipt string) error
    GetOrderItems(orderID uuid.UUID) ([]entity.OrderItem, error)
}

type itemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) ItemRepository {
	return &itemRepository{db: db}
}

func (r *itemRepository) GetShopByUserID(userID uuid.UUID) (*entity.Shop, error) {
	var shop entity.Shop

	query := `
		SELECT id, user_id, name, description, address, created_at, updated_at
		FROM shops
		WHERE user_id = $1
	`

	err := r.db.QueryRow(query, userID).Scan(
		&shop.ID, &shop.UserID, &shop.Name, &shop.Description,
		&shop.Address, &shop.CreatedAt, &shop.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &shop, nil
}

// [internal/repository/postgresql/item_repository.go]

func (r *itemRepository) GetShopOwnerID(shopID uuid.UUID) (uuid.UUID, error) {
    var ownerID uuid.UUID
    query := `SELECT user_id FROM shops WHERE id = $1`
    
    err := r.db.QueryRow(query, shopID).Scan(&ownerID)

    if err == sql.ErrNoRows {
        // Kembalikan uuid.Nil jika toko tidak ditemukan
        return uuid.Nil, errors.New("shop not found")
    }
    if err != nil {
        return uuid.Nil, err
    }
    return ownerID, nil
}

func (r *itemRepository) IsCategoryOwnedByShop(categoryID, shopID uuid.UUID) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1 FROM categories
			WHERE id = $1 AND shop_id = $2
		)
	`

	err := r.db.QueryRow(query, categoryID, shopID).Scan(&exists)
	return exists, err
}

func (r *itemRepository) CreateItem(item *entity.Item) error {
	query := `
		INSERT INTO items (id, shop_id, category_id, name, description, price, stock, condition, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),NOW())
	`

	_, err := r.db.Exec(query,
		item.ID, item.ShopID, item.CategoryID, item.Name,
		item.Description, item.Price, item.Stock, item.Condition,
		item.Status,
	)
	return err
}

func (r *itemRepository) CreateItemImage(img *entity.ItemImage) error {
	query := `
		INSERT INTO item_images (id, item_id, image_url, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := r.db.Exec(query, img.ID, img.ItemID, img.ImageURL)
	return err
}

func (r *itemRepository) GetItemByID(id uuid.UUID) (*entity.Item, error) {
    var item entity.Item
    query := `
        SELECT id, shop_id, category_id, name, description, price, stock, condition, status, created_at, updated_at
        FROM items WHERE id = $1
    `
    err := r.db.QueryRow(query, id).Scan(
        &item.ID, &item.ShopID, &item.CategoryID, &item.Name, &item.Description,
        &item.Price, &item.Stock, &item.Condition, &item.Status, &item.CreatedAt, &item.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, nil // Tidak error, tapi data kosong
    }
    return &item, err
}

func (r *itemRepository) UpdateItem(item *entity.Item) error {
    query := `
        UPDATE items
        SET name=$1, description=$2, price=$3, stock=$4, condition=$5, status=$6, updated_at=NOW()
        WHERE id=$7
    `
    _, err := r.db.Exec(query,
        item.Name, item.Description, item.Price, item.Stock, item.Condition, item.Status, item.ID,
    )
    return err
}

func (r *itemRepository) CreateOffer(offer *entity.Offer) error {
    query := `
        INSERT INTO offers (id, giver_id, seller_id, item_name, description, image_url, expected_price, condition, location, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
    `
    // seller_id harus diubah ke interface{} atau sql.NullUUID jika boleh NULL
    _, err := r.db.Exec(query,
        offer.ID, offer.GiverID, offer.SellerID, offer.ItemName, offer.Description,
        offer.ImageURL, offer.ExpectedPrice, offer.Condition, offer.Location, offer.Status,
    )
    return err
}

// FR-GIVER-03: Melihat Status Penawaran
func (r *itemRepository) GetOffersByGiverID(giverID uuid.UUID) ([]entity.Offer, error) {
    var offers []entity.Offer
    query := `
        SELECT id, giver_id, seller_id, item_name, description, image_url, expected_price, agreed_price, condition, location, status, created_at, updated_at
        FROM offers
        WHERE giver_id = $1
    `
    rows, err := r.db.Query(query, giverID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var offer entity.Offer
        err := rows.Scan(
            &offer.ID, &offer.GiverID, &offer.SellerID, &offer.ItemName, &offer.Description,
            &offer.ImageURL, &offer.ExpectedPrice, &offer.AgreedPrice, &offer.Condition, &offer.Location, &offer.Status, &offer.CreatedAt, &offer.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        offers = append(offers, offer)
    }
    return offers, nil
}

// FR-OFFER-01: Seller Melihat Penawaran
func (r *itemRepository) GetOffersBySellerID(sellerID uuid.UUID) ([]entity.Offer, error) {
    var offers []entity.Offer
    query := `
        SELECT id, giver_id, seller_id, item_name, description, image_url, expected_price, agreed_price, condition, location, status, created_at, updated_at
        FROM offers
        WHERE seller_id = $1 OR seller_id IS NULL -- Jika open offer juga diizinkan dilihat, sesuaikan query ini
    `
    rows, err := r.db.Query(query, sellerID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        // Asumsi struct Offer sudah menggunakan sql.NullFloat64 untuk agreed_price
        var offer entity.Offer
        err := rows.Scan(
            &offer.ID, &offer.GiverID, &offer.SellerID, &offer.ItemName, &offer.Description,
            &offer.ImageURL, &offer.ExpectedPrice, &offer.AgreedPrice, &offer.Condition, &offer.Location, &offer.Status, &offer.CreatedAt, &offer.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        offers = append(offers, offer)
    }
    return offers, nil
}

// Diperlukan untuk mengecek ownership sebelum update status
func (r *itemRepository) GetOfferByID(offerID uuid.UUID) (*entity.Offer, error) {
    var offer entity.Offer
    query := `
        SELECT id, giver_id, seller_id, item_name, description, image_url, expected_price, agreed_price, condition, location, status, created_at, updated_at
        FROM offers WHERE id = $1
    `
    // Asumsi struct Offer sudah menggunakan sql.NullFloat64
    err := r.db.QueryRow(query, offerID).Scan(
        &offer.ID, &offer.GiverID, &offer.SellerID, &offer.ItemName, &offer.Description,
        &offer.ImageURL, &offer.ExpectedPrice, &offer.AgreedPrice, &offer.Condition, &offer.Location, &offer.Status, &offer.CreatedAt, &offer.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &offer, err
}

// FR-OFFER-02/03: Update Status dan Agreed Price
func (r *itemRepository) UpdateOffer(offer *entity.Offer) error {
    query := `
        UPDATE offers
        SET status=$1, agreed_price=$2, updated_at=NOW()
        WHERE id=$3
    `
    // agreed_price (offer.AgreedPrice) sekarang bertipe sql.NullFloat64
    _, err := r.db.Exec(query, offer.Status, offer.AgreedPrice, offer.ID)
    return err
}

// Implementasi
// FR-BUYER-01 & FR-BUYER-02: Melihat & Filter Marketplace
func (r *itemRepository) GetMarketItems(filter entity.ItemFilter) ([]entity.Item, error) {
    // Implementasi ini akan panjang karena harus membuat query dinamis
    // Untuk mempersingkat, kita buat query dasar saja
    var items []entity.Item
    
    // Query dasar: status='active' dan stock > 0 (FR-BUYER-01)
    query := `
        SELECT id, shop_id, category_id, name, description, price, stock, condition, status, created_at, updated_at
        FROM items
        WHERE status = 'active' AND stock > 0
    `
    // Implementasi filter (FR-BUYER-02) di sini: Keyword, CategoryID, Price Range, Limit, Offset.
    // ... logic pembuatan WHERE clause dinamis ... 
    
    // Karena ini contoh, kita eksekusi query dasar tanpa filter dinamis:
    rows, err := r.db.Query(query) 
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var item entity.Item
        // Scan semua field
        // ...
        items = append(items, item)
    }
    return items, nil
}

// FR-BUYER-03 & FR-BUYER-04: Ambil Item Detail
func (r *itemRepository) GetItemForOrder(itemID uuid.UUID) (*entity.Item, error) {
    // Query sama dengan GetItemByID, tapi mungkin ditambahkan JOIN ke Shop
    // Untuk mempersingkat, kita gunakan GetItemByID yang sudah ada
    // ...
    // Pastikan item ada, aktif, dan stok > 0
    // ...
    return r.GetItemByID(itemID) // Asumsi sudah cukup
}

// FR-BUYER-04: Membuat Order (Menggunakan Transaksi)
func (r *itemRepository) CreateOrderTransaction(order *entity.Order, orderItems []entity.OrderItem) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    
    // 1. Insert Order
    orderQuery := `
        INSERT INTO orders (id, buyer_id, shop_id, total_price, status, shipping_address, shipping_courier, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
    `
    if _, err := tx.Exec(orderQuery, order.ID, order.BuyerID, order.ShopID, order.TotalPrice, order.Status, order.ShippingAddress, order.ShippingCourier); err != nil {
        tx.Rollback()
        return err
    }

    // 2. Insert Order Items & Update Stock (Loop)
    for _, item := range orderItems {
        // Insert Order Item
        itemQuery := `INSERT INTO order_items (id, order_id, item_id, quantity, price, created_at) VALUES ($1, $2, $3, $4, $5, NOW())`
        if _, err := tx.Exec(itemQuery, uuid.New(), item.OrderID, item.ItemID, item.Quantity, item.Price); err != nil {
            tx.Rollback()
            return err
        }
        
        // Update Stock (Decrement)
        stockQuery := `UPDATE items SET stock = stock - $1 WHERE id = $2`
        if _, err := tx.Exec(stockQuery, item.Quantity, item.ItemID); err != nil {
            tx.Rollback()
            return err
        }
    }

    return tx.Commit()
}

func (r *itemRepository) GetOrderByID(orderID uuid.UUID) (*entity.Order, error) {
    var order entity.Order
    query := `
        SELECT id, buyer_id, shop_id, total_price, status, shipping_address, shipping_courier, shipping_receipt, created_at, updated_at
        FROM orders WHERE id = $1
    `
    // Implementasi Scan...
    err := r.db.QueryRow(query, orderID).Scan(
        &order.ID, &order.BuyerID, &order.ShopID, &order.TotalPrice, &order.Status, 
        &order.ShippingAddress, &order.ShippingCourier, &order.ShippingReceipt, &order.CreatedAt, &order.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &order, err
}

// FR-ORDER-02: Update Status
func (r *itemRepository) UpdateOrderStatus(orderID uuid.UUID, status string) error {
    query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
    _, err := r.db.Exec(query, status, orderID)
    return err
}

// FR-ORDER-03: Input Nomor Resi
func (r *itemRepository) UpdateOrderShipment(orderID uuid.UUID, courier string, receipt string) error {
    // FR-ORDER-03 juga mengubah status menjadi 'shipped'
    query := `
        UPDATE orders SET shipping_courier = $1, shipping_receipt = $2, status = 'shipped', updated_at = NOW() 
        WHERE id = $3
    `
    _, err := r.db.Exec(query, courier, receipt, orderID)
    return err
}

// FR-ORDER-04: Ambil Order Items untuk Tracking
func (r *itemRepository) GetOrderItems(orderID uuid.UUID) ([]entity.OrderItem, error) {
    // Implementasi query SELECT dari tabel order_items
    // ...
    return nil, nil // Placeholder
}