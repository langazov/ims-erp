package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ProductStatus string

const (
	ProductStatusDraft        ProductStatus = "draft"
	ProductStatusActive       ProductStatus = "active"
	ProductStatusInactive     ProductStatus = "inactive"
	ProductStatusDiscontinued ProductStatus = "discontinued"
)

type ProductType string

const (
	ProductTypeGood         ProductType = "good"
	ProductTypeService      ProductType = "service"
	ProductTypeSubscription ProductType = "subscription"
)

type ProductCategory string

const (
	CategoryRawMaterial  ProductCategory = "raw_material"
	CategoryFinishedGood ProductCategory = "finished_good"
	CategoryComponent    ProductCategory = "component"
	CategoryPackaging    ProductCategory = "packaging"
	CategoryService      ProductCategory = "service"
)

type Product struct {
	ID               uuid.UUID       `json:"id" bson:"_id"`
	TenantID         uuid.UUID       `json:"tenantId" bson:"tenantId"`
	SKU              string          `json:"sku" bson:"sku"`
	Barcode          string          `json:"barcode" bson:"barcode"`
	Name             string          `json:"name" bson:"name"`
	Description      string          `json:"description" bson:"description"`
	ShortDescription string          `json:"shortDescription" bson:"shortDescription"`
	Type             ProductType     `json:"type" bson:"type"`
	Category         ProductCategory `json:"category" bson:"category"`
	Status           ProductStatus   `json:"status" bson:"status"`
	Currency         string          `json:"currency" bson:"currency"`
	Brand            string          `json:"brand" bson:"brand"`
	Manufacturer     string          `json:"manufacturer" bson:"manufacturer"`

	Pricing ProductPricing  `json:"pricing" bson:"pricing"`
	Cost    decimal.Decimal `json:"cost" bson:"cost"`

	Inventory ProductInventory `json:"inventory" bson:"inventory"`

	Weight     decimal.Decimal `json:"weight" bson:"weight"`
	WeightUnit string          `json:"weightUnit" bson:"weightUnit"`
	Dimensions Dimensions      `json:"dimensions" bson:"dimensions"`

	Images    []ProductImage    `json:"images" bson:"images"`
	Documents []ProductDocument `json:"documents" bson:"documents"`

	Attributes map[string]interface{} `json:"attributes" bson:"attributes"`
	Tags       []string               `json:"tags" bson:"tags"`
	Metadata   map[string]string      `json:"metadata" bson:"metadata"`

	VariantOf *uuid.UUID  `json:"variantOf" bson:"variantOf"`
	Variants  []uuid.UUID `json:"variants" bson:"variants"`

	SalesChannels []string `json:"salesChannels" bson:"salesChannels"`
	Channels      []string `json:"channels" bson:"channels"`

	TaxCategory string `json:"taxCategory" bson:"taxCategory"`
	HSNCode     string `json:"hsnCode" bson:"hsnCode"`

	CreatedBy uuid.UUID `json:"createdBy" bson:"createdBy"`
	UpdatedBy uuid.UUID `json:"updatedBy" bson:"updatedBy"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
	Version   int64     `json:"-" bson:"version"`
}

type ProductPricing struct {
	ListPrice      decimal.Decimal `json:"listPrice" bson:"listPrice"`
	SalePrice      decimal.Decimal `json:"salePrice" bson:"salePrice"`
	Currency       string          `json:"currency" bson:"currency"`
	CostPrice      decimal.Decimal `json:"costPrice" bson:"costPrice"`
	Margin         decimal.Decimal `json:"margin" bson:"margin"`
	MarginPercent  decimal.Decimal `json:"marginPercent" bson:"marginPercent"`
	MSRP           decimal.Decimal `json:"msrp" bson:"msrp"`
	WholeSalePrice decimal.Decimal `json:"wholeSalePrice" bson:"wholeSalePrice"`
	MinPrice       decimal.Decimal `json:"minPrice" bson:"minPrice"`
	MaxPrice       decimal.Decimal `json:"maxPrice" bson:"maxPrice"`
	Discount       decimal.Decimal `json:"discount" bson:"discount"`
	DiscountType   string          `json:"discountType" bson:"discountType"`
	ValidFrom      *time.Time      `json:"validFrom" bson:"validFrom"`
	ValidUntil     *time.Time      `json:"validUntil" bson:"validUntil"`
}

type ProductInventory struct {
	TrackInventory    bool   `json:"trackInventory" bson:"trackInventory"`
	QuantityOnHand    int    `json:"quantityOnHand" bson:"quantityOnHand"`
	QuantityReserved  int    `json:"quantityReserved" bson:"quantityReserved"`
	QuantityAvailable int    `json:"quantityAvailable" bson:"quantityAvailable"`
	ReorderPoint      int    `json:"reorderPoint" bson:"reorderPoint"`
	ReorderQuantity   int    `json:"reorderQuantity" bson:"reorderQuantity"`
	LeadTime          int    `json:"leadTime" bson:"leadTime"`
	StockLocation     string `json:"stockLocation" bson:"stockLocation"`
	BinLocation       string `json:"binLocation" bson:"binLocation"`
	BatchTracking     bool   `json:"batchTracking" bson:"batchTracking"`
	SerialTracking    bool   `json:"serialTracking" bson:"serialTracking"`
	AllowBackorder    bool   `json:"allowBackorder" bson:"allowBackorder"`
	MinStockLevel     int    `json:"minStockLevel" bson:"minStockLevel"`
	MaxStockLevel     int    `json:"maxStockLevel" bson:"maxStockLevel"`
}

type Dimensions struct {
	Length decimal.Decimal `json:"length" bson:"length"`
	Width  decimal.Decimal `json:"width" bson:"width"`
	Height decimal.Decimal `json:"height" bson:"height"`
	Unit   string          `json:"unit" bson:"unit"`
}

type ProductImage struct {
	ID        uuid.UUID `json:"id" bson:"id"`
	URL       string    `json:"url" bson:"url"`
	AltText   string    `json:"altText" bson:"altText"`
	Position  int       `json:"position" bson:"position"`
	IsPrimary bool      `json:"isPrimary" bson:"isPrimary"`
}

type ProductDocument struct {
	ID         uuid.UUID `json:"id" bson:"id"`
	Name       string    `json:"name" bson:"name"`
	Type       string    `json:"type" bson:"type"`
	URL        string    `json:"url" bson:"url"`
	Size       int64     `json:"size" bson:"size"`
	UploadedAt time.Time `json:"uploadedAt" bson:"uploadedAt"`
}

func NewProduct(
	tenantID, createdBy uuid.UUID,
	sku, name string,
	productType ProductType,
	category ProductCategory,
	currency string,
) (*Product, error) {
	now := time.Now().UTC()
	id := uuid.New()

	product := &Product{
		ID:       id,
		TenantID: tenantID,
		SKU:      sku,
		Name:     name,
		Type:     productType,
		Category: category,
		Status:   ProductStatusDraft,
		Pricing: ProductPricing{
			Currency: currency,
		},
		Inventory: ProductInventory{
			TrackInventory: true,
		},
		Attributes:    make(map[string]interface{}),
		Tags:          []string{},
		Metadata:      make(map[string]string),
		Images:        []ProductImage{},
		Documents:     []ProductDocument{},
		Variants:      []uuid.UUID{},
		SalesChannels: []string{},
		Channels:      []string{},
		CreatedBy:     createdBy,
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       0,
	}

	return product, nil
}

func (p *Product) SetName(name string) {
	p.Name = name
	p.UpdatedAt = time.Now().UTC()
}

func (p *Product) SetDescription(description string) {
	p.Description = description
	p.UpdatedAt = time.Now().UTC()
}

func (p *Product) SetPricing(listPrice, salePrice, costPrice decimal.Decimal) {
	p.Pricing.ListPrice = listPrice
	p.Pricing.SalePrice = salePrice
	p.Pricing.CostPrice = costPrice
	p.Pricing.Margin = listPrice.Sub(costPrice)
	if !costPrice.IsZero() {
		p.Pricing.MarginPercent = p.Pricing.Margin.Div(costPrice).Mul(decimal.NewFromInt(100))
	}
	p.UpdatedAt = time.Now().UTC()
}

func (p *Product) SetInventory(quantityOnHand, reorderPoint, reorderQuantity int) {
	p.Inventory.QuantityOnHand = quantityOnHand
	p.Inventory.ReorderPoint = reorderPoint
	p.Inventory.ReorderQuantity = reorderQuantity
	p.Inventory.QuantityAvailable = quantityOnHand - p.Inventory.QuantityReserved
	p.UpdatedAt = time.Now().UTC()
}

func (p *Product) AdjustInventory(adjustment int, reason string) {
	p.Inventory.QuantityOnHand += adjustment
	p.Inventory.QuantityAvailable = p.Inventory.QuantityOnHand - p.Inventory.QuantityReserved
	p.UpdatedAt = time.Now().UTC()
}

func (p *Product) ReserveStock(quantity int) error {
	if quantity > p.Inventory.QuantityAvailable {
		return ErrInsufficientStock
	}
	p.Inventory.QuantityReserved += quantity
	p.Inventory.QuantityAvailable -= quantity
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *Product) ReleaseReservation(quantity int) {
	p.Inventory.QuantityReserved -= quantity
	p.Inventory.QuantityAvailable = p.Inventory.QuantityOnHand - p.Inventory.QuantityReserved
	p.UpdatedAt = time.Now().UTC()
}

func (p *Product) CommitReservation(quantity int) {
	p.Inventory.QuantityOnHand -= quantity
	p.Inventory.QuantityReserved -= quantity
	p.Inventory.QuantityAvailable = p.Inventory.QuantityOnHand - p.Inventory.QuantityReserved
	p.UpdatedAt = time.Now().UTC()
}

func (p *Product) Activate() {
	if p.Status == ProductStatusDraft || p.Status == ProductStatusInactive {
		p.Status = ProductStatusActive
		p.UpdatedAt = time.Now().UTC()
	}
}

func (p *Product) Deactivate() {
	if p.Status == ProductStatusActive {
		p.Status = ProductStatusInactive
		p.UpdatedAt = time.Now().UTC()
	}
}

func (p *Product) Discontinue() {
	p.Status = ProductStatusDiscontinued
	p.UpdatedAt = time.Now().UTC()
}

func (p *Product) AddImage(image ProductImage) {
	image.ID = uuid.New()
	image.Position = len(p.Images) + 1
	p.Images = append(p.Images, image)
	p.UpdatedAt = time.Now().UTC()
}

func (p *Product) RemoveImage(imageID uuid.UUID) {
	newImages := make([]ProductImage, 0, len(p.Images))
	for _, img := range p.Images {
		if img.ID != imageID {
			newImages = append(newImages, img)
		}
	}
	p.Images = newImages
	p.UpdatedAt = time.Now().UTC()
}

func (p *Product) SetPrimaryImage(imageID uuid.UUID) {
	for i := range p.Images {
		p.Images[i].IsPrimary = p.Images[i].ID == imageID
	}
	p.UpdatedAt = time.Now().UTC()
}

func (p *Product) AddTag(tag string) {
	p.Tags = append(p.Tags, tag)
	p.UpdatedAt = time.Now().UTC()
}

func (p *Product) RemoveTag(tag string) {
	newTags := make([]string, 0, len(p.Tags))
	for _, t := range p.Tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}
	p.Tags = newTags
	p.UpdatedAt = time.Now().UTC()
}

func (p *Product) AddVariant(variantID uuid.UUID) {
	p.Variants = append(p.Variants, variantID)
	p.UpdatedAt = time.Now().UTC()
}

func (p *Product) SetAttribute(key string, value interface{}) {
	if p.Attributes == nil {
		p.Attributes = make(map[string]interface{})
	}
	p.Attributes[key] = value
	p.UpdatedAt = time.Now().UTC()
}

func (p *Product) GetStockStatus() StockStatus {
	if !p.Inventory.TrackInventory {
		return StockStatusAvailable
	}
	if p.Inventory.QuantityOnHand <= 0 {
		if p.Inventory.AllowBackorder {
			return StockStatusBackorder
		}
		return StockStatusOutOfStock
	}
	if p.Inventory.QuantityOnHand <= p.Inventory.ReorderPoint {
		return StockStatusLowStock
	}
	return StockStatusAvailable
}

type StockStatus string

const (
	StockStatusAvailable  StockStatus = "available"
	StockStatusLowStock   StockStatus = "low_stock"
	StockStatusOutOfStock StockStatus = "out_of_stock"
	StockStatusBackorder  StockStatus = "backorder"
)

var ErrInsufficientStock = &ProductError{
	Code:    "INSUFFICIENT_STOCK",
	Message: "Insufficient stock available",
}

type ProductError struct {
	Code    string
	Message string
}

func (e *ProductError) Error() string {
	return e.Message
}
