package handler

import (
	entity "home-market/internal/domain"
	service "home-market/internal/service/postgresql"
	// "mime/multipart"
	// "net/http"
	"fmt"
	// "io"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ItemHandler struct {
	itemService *service.ItemService
}

func NewItemHandler(itemService *service.ItemService) *ItemHandler {
	return &ItemHandler{itemService: itemService}
}

func (h *ItemHandler) CreateItem(c *gin.Context) {
	fmt.Println("Content-Type:", c.GetHeader("Content-Type"))
	// body, _ := io.ReadAll(c.Request.Body)
	// fmt.Println("BODY LENGTH:", len(body))

	userID := c.MustGet("user_id").(uuid.UUID)
	role := c.MustGet("role_name").(string)

	// --- FORM MULTIPART ---
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid form-data", "detail": err.Error()})
		return
	}

	// helper
	get := func(key string) string {
		if v, ok := form.Value[key]; ok && len(v) > 0 {
			return v[0]
		}
		return ""
	}

	// --- Ambil text fields ---
	name := get("name")
	description := get("description")
	priceStr := get("price")
	stockStr := get("stock")
	condition := get("condition")
	categoryIDStr := get("category_id")

	if name == "" || priceStr == "" || stockStr == "" || categoryIDStr == "" {
		c.JSON(400, gin.H{"error": "missing required fields"})
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid price"})
		return
	}

	stock, err := strconv.Atoi(stockStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid stock"})
		return
	}

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid category_id"})
		return
	}

	input := entity.CreateItemInput{
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		Condition:   condition,
		CategoryID:  categoryID,
	}

	// --- Images ---
	files := form.File["images"]

	var imageURLs []string
	for _, file := range files {
		filename := uuid.New().String() + filepath.Ext(file.Filename)
		savePath := "uploads/items/" + filename

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		imageURLs = append(imageURLs, "/uploads/items/"+filename)
	}

	// --- Service ---
	item, images, err := h.itemService.CreateItem(userID, role, input, imageURLs)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"item":   item,
		"images": images,
	})
}

func (h *ItemHandler) UpdateItem(c *gin.Context) {
    idStr := c.Param("id")
    itemID, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(400, gin.H{"error": "invalid item id"})
        return
    }

    var input entity.UpdateItemInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, gin.H{"error": "invalid input", "detail": err.Error()})
        return
    }

    userID := c.MustGet("user_id").(uuid.UUID)

    updatedItem, err := h.itemService.UpdateItem(userID, itemID, input)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"message": "item updated", "data": updatedItem})
}

func (h *ItemHandler) DeleteItem(c *gin.Context) {
    idStr := c.Param("id")
    itemID, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(400, gin.H{"error": "invalid item id"})
        return
    }

    userID := c.MustGet("user_id").(uuid.UUID)

    if err := h.itemService.DeleteItem(userID, itemID); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"message": "item archived/deleted successfully"})
}

// FR-GIVER-01 & FR-GIVER-02: Membuat Penawaran
func (h *ItemHandler) CreateOffer(c *gin.Context) {
    userID := c.MustGet("user_id").(uuid.UUID)
    role := c.MustGet("role_name").(string)

    // Cek Role sebelum Parsing (efisiensi)
    if role != "giver" {
        c.JSON(403, gin.H{"error": "Forbidden: only Giver can create offers"})
        return
    }

    // Menggunakan c.MultipartForm() untuk handle form-data + file
    form, err := c.MultipartForm()
    if err != nil {
        c.JSON(400, gin.H{"error": "invalid form-data", "detail": err.Error()})
        return
    }

    // Helper untuk mengambil nilai form
    get := func(key string) string {
        if v, ok := form.Value[key]; ok && len(v) > 0 {
            return v[0]
        }
        return ""
    }

    // Mapping input manual dari form
    input := entity.CreateOfferInput{
        SellerIDStr:   get("seller_id"),
        ItemName:      get("item_name"),
        Description:   get("description"),
        Condition:     get("condition"),
        Location:      get("location"),
    }
    
    // Parse Expected Price
    priceStr := get("expected_price")
    input.ExpectedPrice, err = strconv.ParseFloat(priceStr, 64)
    if err != nil {
        c.JSON(400, gin.H{"error": "invalid expected_price"})
        return
    }

    // --- Images Upload (FR-GIVER-02) ---
    files := form.File["images"]
    if len(files) == 0 {
        c.JSON(400, gin.H{"error": "image file is required"})
        return
    }
    
    // Ambil hanya 1 file (asumsi SRS hanya butuh 1 image_url)
    file := files[0]
    filename := uuid.New().String() + filepath.Ext(file.Filename)
    savePath := "uploads/offers/" + filename // Simpan di folder berbeda
    
    if err := c.SaveUploadedFile(file, savePath); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    imageURL := "/uploads/offers/" + filename
    
    // --- Service Call ---
    offer, err := h.itemService.CreateOffer(userID, role, input, imageURL)
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, gin.H{
        "message": "Offer created successfully. Waiting for seller response.",
        "offer": offer,
    })
}

// FR-GIVER-03: Melihat Status Penawaran
func (h *ItemHandler) GetMyOffers(c *gin.Context) {
    userID := c.MustGet("user_id").(uuid.UUID)
    role := c.MustGet("role_name").(string)

    // Cek Role
    if role != "giver" {
        c.JSON(403, gin.H{"error": "Forbidden: only Giver can view offers"})
        return
    }

    offers, err := h.itemService.GetMyOffers(userID, role)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"offers": offers})
}

// [internal/delivery/http/handler/item_handler.go]

// FR-OFFER-01: Seller Melihat Penawaran
func (h *ItemHandler) GetOffersToSeller(c *gin.Context) {
    userID := c.MustGet("user_id").(uuid.UUID)
    role := c.MustGet("role_name").(string)

    offers, err := h.itemService.GetOffersToSeller(userID, role)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"offers": offers})
}

// FR-OFFER-02: Seller Menerima Penawaran
func (h *ItemHandler) AcceptOffer(c *gin.Context) {
    offerIDStr := c.Param("id")
    offerID, err := uuid.Parse(offerIDStr)
    if err != nil {
        c.JSON(400, gin.H{"error": "invalid offer id"})
        return
    }

    var input entity.AcceptOfferInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, gin.H{"error": "invalid input for agreed price", "detail": err.Error()})
        return
    }

    userID := c.MustGet("user_id").(uuid.UUID)

    offer, draftItem, err := h.itemService.AcceptOffer(userID, offerID, input)
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "message": "Offer accepted successfully. Draft item created.",
        "offer": offer,
        "draft_item": draftItem, // FR-OFFER-04
    })
}

// FR-OFFER-03: Seller Menolak Penawaran
func (h *ItemHandler) RejectOffer(c *gin.Context) {
    offerIDStr := c.Param("id")
    offerID, err := uuid.Parse(offerIDStr)
    if err != nil {
        c.JSON(400, gin.H{"error": "invalid offer id"})
        return
    }

    userID := c.MustGet("user_id").(uuid.UUID)

    offer, err := h.itemService.RejectOffer(userID, offerID)
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "message": "Offer rejected successfully.",
        "offer": offer,
    })
}

// [internal/delivery/http/handler/item_handler.go]

// FR-BUYER-01 & FR-BUYER-02: Melihat & Filter Marketplace
func (h *ItemHandler) GetMarketplaceItems(c *gin.Context) {
    var filter entity.ItemFilter
    // c.ShouldBindQuery dapat menangani ItemFilter
    if err := c.ShouldBindQuery(&filter); err != nil {
        c.JSON(400, gin.H{"error": "invalid query parameters"})
        return
    }

    items, err := h.itemService.GetMarketplaceItems(filter)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"data": items})
}

// FR-BUYER-03: Melihat Detail Barang
func (h *ItemHandler) GetItemDetail(c *gin.Context) {
    idStr := c.Param("id")
    itemID, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(400, gin.H{"error": "invalid item id"})
        return
    }

    item, err := h.itemService.GetItemDetail(itemID)
    if err != nil {
        c.JSON(404, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"data": item})
}

// FR-BUYER-04: Membuat Order
func (h *ItemHandler) CreateOrder(c *gin.Context) {
    userID := c.MustGet("user_id").(uuid.UUID)
    role := c.MustGet("role_name").(string)
    
    if role != "buyer" {
        c.JSON(403, gin.H{"error": "Forbidden: only Buyer can create orders"})
        return
    }

    var input entity.CreateOrderInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, gin.H{"error": "invalid input", "detail": err.Error()})
        return
    }

    order, err := h.itemService.CreateOrder(userID, input)
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, gin.H{"message": "Order created successfully", "order": order})
}

func (h *ItemHandler) UpdateOrderStatus(c *gin.Context) {
    // Implementasi dari service.UpdateOrderStatus (FR-ORDER-02)
    // Lihat panduan implementasi di jawaban sebelumnya.
}

func (h *ItemHandler) InputShippingReceipt(c *gin.Context) {
    // Implementasi dari service.InputShippingReceipt (FR-ORDER-03)
    // Lihat panduan implementasi di jawaban sebelumnya.
}

func (h *ItemHandler) GetOrderTracking(c *gin.Context) {
    // Implementasi dari service.GetOrderTracking (FR-ORDER-04)
    // Lihat panduan implementasi di jawaban sebelumnya.
}