package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProduct(t *testing.T) {
	tenantID := uuid.New()
	createdBy := uuid.New()
	sku := "SKU-001"
	name := "Test Product"

	product, err := NewProduct(tenantID, createdBy, sku, name, ProductTypeGood, CategoryFinishedGood, "USD")

	require.NoError(t, err)
	assert.NotEmpty(t, product.ID)
	assert.Equal(t, tenantID, product.TenantID)
	assert.Equal(t, sku, product.SKU)
	assert.Equal(t, name, product.Name)
	assert.Equal(t, ProductTypeGood, product.Type)
	assert.Equal(t, CategoryFinishedGood, product.Category)
	assert.Equal(t, ProductStatusDraft, product.Status)
	assert.Equal(t, "USD", product.Pricing.Currency)
	assert.False(t, product.CreatedAt.IsZero())
}

func TestProductSetName(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Original Name", ProductTypeGood, CategoryFinishedGood, "USD")

	product.SetName("Updated Name")

	assert.Equal(t, "Updated Name", product.Name)
}

func TestProductSetDescription(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")

	product.SetDescription("Updated description")

	assert.Equal(t, "Updated description", product.Description)
}

func TestProductSetPricing(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")

	product.SetPricing(decimal.NewFromFloat(100.00), decimal.NewFromFloat(80.00), decimal.NewFromFloat(70.00))

	assert.True(t, product.Pricing.ListPrice.Equal(decimal.NewFromFloat(100.00)))
	assert.True(t, product.Pricing.SalePrice.Equal(decimal.NewFromFloat(80.00)))
	assert.True(t, product.Pricing.CostPrice.Equal(decimal.NewFromFloat(70.00)))
	assert.True(t, product.Pricing.Margin.Equal(decimal.NewFromFloat(30.00)))
}

func TestProductSetInventory(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")

	product.SetInventory(100, 10, 50)

	assert.Equal(t, 100, product.Inventory.QuantityOnHand)
	assert.Equal(t, 10, product.Inventory.ReorderPoint)
	assert.Equal(t, 50, product.Inventory.ReorderQuantity)
}

func TestProductAdjustInventory(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")
	product.SetInventory(100, 10, 50)

	product.AdjustInventory(25, "Stock received")

	assert.Equal(t, 125, product.Inventory.QuantityOnHand)
}

func TestProductReserveStock(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")
	product.SetInventory(100, 10, 50)

	err := product.ReserveStock(25)

	require.NoError(t, err)
	assert.Equal(t, 75, product.Inventory.QuantityAvailable)
	assert.Equal(t, 25, product.Inventory.QuantityReserved)
}

func TestProductReserveStockInsufficient(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")
	product.SetInventory(100, 10, 50)

	err := product.ReserveStock(150)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Insufficient stock")
}

func TestProductReleaseReservation(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")
	product.SetInventory(100, 10, 75)
	product.Inventory.QuantityReserved = 25

	product.ReleaseReservation(10)

	assert.Equal(t, 85, product.Inventory.QuantityAvailable)
	assert.Equal(t, 15, product.Inventory.QuantityReserved)
}

func TestProductCommitReservation(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")
	product.SetInventory(100, 10, 75)
	product.Inventory.QuantityReserved = 25

	product.CommitReservation(20)

	assert.Equal(t, 80, product.Inventory.QuantityOnHand)
	assert.Equal(t, 5, product.Inventory.QuantityReserved)
}

func TestProductActivate(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")

	assert.Equal(t, ProductStatusDraft, product.Status)

	product.Activate()

	assert.Equal(t, ProductStatusActive, product.Status)
}

func TestProductDeactivate(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")
	product.Activate()

	product.Deactivate()

	assert.Equal(t, ProductStatusInactive, product.Status)
}

func TestProductDiscontinue(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")

	product.Discontinue()

	assert.Equal(t, ProductStatusDiscontinued, product.Status)
}

func TestProductAddImage(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")

	image := ProductImage{
		URL:     "https://example.com/image.jpg",
		AltText: "Product image",
	}
	product.AddImage(image)

	require.Len(t, product.Images, 1)
	assert.Equal(t, "Product image", product.Images[0].AltText)
	assert.NotEmpty(t, product.Images[0].ID)
}

func TestProductRemoveImage(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")

	image := ProductImage{
		URL: "https://example.com/image.jpg",
	}
	product.AddImage(image)
	imageID := product.Images[0].ID

	product.RemoveImage(imageID)

	assert.Empty(t, product.Images)
}

func TestProductSetPrimaryImage(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")

	image1 := ProductImage{URL: "https://example.com/image1.jpg"}
	image2 := ProductImage{URL: "https://example.com/image2.jpg"}
	product.AddImage(image1)
	product.AddImage(image2)
	image2ID := product.Images[1].ID

	product.SetPrimaryImage(image2ID)

	assert.False(t, product.Images[0].IsPrimary)
	assert.True(t, product.Images[1].IsPrimary)
}

func TestProductAddTag(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")

	product.AddTag("electronics")
	product.AddTag("premium")

	require.Len(t, product.Tags, 2)
	assert.Contains(t, product.Tags, "electronics")
	assert.Contains(t, product.Tags, "premium")
}

func TestProductRemoveTag(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")
	product.AddTag("electronics")
	product.AddTag("premium")

	product.RemoveTag("electronics")

	assert.Len(t, product.Tags, 1)
	assert.Contains(t, product.Tags, "premium")
}

func TestProductAddVariant(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")
	variantID := uuid.New()

	product.AddVariant(variantID)

	assert.Contains(t, product.Variants, variantID)
}

func TestProductSetAttribute(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")

	product.SetAttribute("color", "red")
	product.SetAttribute("size", "large")

	color, ok := product.Attributes["color"]
	assert.True(t, ok)
	assert.Equal(t, "red", color)
}

func TestProductGetStockStatusAvailable(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")
	product.SetInventory(100, 10, 50)

	status := product.GetStockStatus()

	assert.Equal(t, StockStatusAvailable, status)
}

func TestProductGetStockStatusLowStock(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")
	product.SetInventory(5, 10, 50)

	status := product.GetStockStatus()

	assert.Equal(t, StockStatusLowStock, status)
}

func TestProductGetStockStatusOutOfStock(t *testing.T) {
	product, _ := NewProduct(uuid.New(), uuid.New(), "SKU-001", "Test Product", ProductTypeGood, CategoryFinishedGood, "USD")
	product.SetInventory(0, 10, 50)

	status := product.GetStockStatus()

	assert.Equal(t, StockStatusOutOfStock, status)
}

func TestProductError(t *testing.T) {
	err := &ProductError{
		Code:    "INSUFFICIENT_STOCK",
		Message: "Not enough stock",
	}

	assert.Equal(t, "INSUFFICIENT_STOCK", err.Code)
	assert.Equal(t, "Not enough stock", err.Error())
}

func TestStockStatusConstants(t *testing.T) {
	assert.Equal(t, StockStatus("available"), StockStatusAvailable)
	assert.Equal(t, StockStatus("low_stock"), StockStatusLowStock)
	assert.Equal(t, StockStatus("out_of_stock"), StockStatusOutOfStock)
	assert.Equal(t, StockStatus("backorder"), StockStatusBackorder)
}

func TestProductTypeConstants(t *testing.T) {
	assert.Equal(t, ProductType("good"), ProductTypeGood)
	assert.Equal(t, ProductType("service"), ProductTypeService)
	assert.Equal(t, ProductType("subscription"), ProductTypeSubscription)
}

func TestProductCategoryConstants(t *testing.T) {
	assert.Equal(t, ProductCategory("raw_material"), CategoryRawMaterial)
	assert.Equal(t, ProductCategory("finished_good"), CategoryFinishedGood)
	assert.Equal(t, ProductCategory("component"), CategoryComponent)
	assert.Equal(t, ProductCategory("packaging"), CategoryPackaging)
	assert.Equal(t, ProductCategory("service"), CategoryService)
}

func TestProductStatusConstants(t *testing.T) {
	assert.Equal(t, ProductStatus("draft"), ProductStatusDraft)
	assert.Equal(t, ProductStatus("active"), ProductStatusActive)
	assert.Equal(t, ProductStatus("inactive"), ProductStatusInactive)
	assert.Equal(t, ProductStatus("discontinued"), ProductStatusDiscontinued)
}
