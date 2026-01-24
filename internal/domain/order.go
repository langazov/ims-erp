package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	OrderStatusDraft             OrderStatus = "draft"
	OrderStatusPending           OrderStatus = "pending"
	OrderStatusConfirmed         OrderStatus = "confirmed"
	OrderStatusProcessing        OrderStatus = "processing"
	OrderStatusShipped           OrderStatus = "shipped"
	OrderStatusDelivered         OrderStatus = "delivered"
	OrderStatusCompleted         OrderStatus = "completed"
	OrderStatusCancelled         OrderStatus = "cancelled"
	OrderStatusRefunded          OrderStatus = "refunded"
	OrderStatusPartiallyRefunded OrderStatus = "partially_refunded"
)

type OrderType string

const (
	OrderTypeStandard     OrderType = "standard"
	OrderTypePreOrder     OrderType = "pre_order"
	OrderTypeSubscription OrderType = "subscription"
	OrderTypeQuote        OrderType = "quote"
)

type OrderSource string

const (
	OrderSourceWeb         OrderSource = "web"
	OrderSourceMobile      OrderSource = "mobile"
	OrderSourceAPI         OrderSource = "api"
	OrderSourcePhone       OrderSource = "phone"
	OrderSourceInStore     OrderSource = "in_store"
	OrderSourceMarketplace OrderSource = "marketplace"
)

type OrderPaymentStatus string

const (
	OrderPaymentStatusPending   OrderPaymentStatus = "pending"
	OrderPaymentStatusPaid      OrderPaymentStatus = "paid"
	OrderPaymentStatusPartially OrderPaymentStatus = "partially_paid"
	OrderPaymentStatusRefunded  OrderPaymentStatus = "refunded"
	OrderPaymentStatusOverpaid  OrderPaymentStatus = "overpaid"
	OrderPaymentStatusFailed    OrderPaymentStatus = "failed"
)

type FulfillmentStatus string

const (
	FulfillmentStatusUnfulfilled FulfillmentStatus = "unfulfilled"
	FulfillmentStatusPartial     FulfillmentStatus = "partial"
	FulfillmentStatusFulfilled   FulfillmentStatus = "fulfilled"
	FulfillmentStatusReturned    FulfillmentStatus = "returned"
)

type Order struct {
	ID                uuid.UUID          `json:"id" bson:"_id"`
	TenantID          uuid.UUID          `json:"tenantId" bson:"tenantId"`
	OrderNumber       string             `json:"orderNumber" bson:"orderNumber"`
	ClientID          uuid.UUID          `json:"clientId" bson:"clientId"`
	Type              OrderType          `json:"type" bson:"type"`
	Source            OrderSource        `json:"source" bson:"source"`
	Status            OrderStatus        `json:"status" bson:"status"`
	PaymentStatus     OrderPaymentStatus `json:"paymentStatus" bson:"paymentStatus"`
	FulfillmentStatus FulfillmentStatus  `json:"fulfillmentStatus" bson:"fulfillmentStatus"`

	Currency      string          `json:"currency" bson:"currency"`
	Subtotal      decimal.Decimal `json:"subtotal" bson:"subtotal"`
	DiscountTotal decimal.Decimal `json:"discountTotal" bson:"discountTotal"`
	TaxTotal      decimal.Decimal `json:"taxTotal" bson:"taxTotal"`
	ShippingTotal decimal.Decimal `json:"shippingTotal" bson:"shippingTotal"`
	HandlingTotal decimal.Decimal `json:"handlingTotal" bson:"handlingTotal"`
	Total         decimal.Decimal `json:"total" bson:"total"`
	AmountPaid    decimal.Decimal `json:"amountPaid" bson:"amountPaid"`
	AmountDue     decimal.Decimal `json:"amountDue" bson:"amountDue"`

	Lines []OrderLine `json:"lines" bson:"lines"`

	BillingAddress   *Address   `json:"billingAddress" bson:"billingAddress"`
	ShippingAddress  *Address   `json:"shippingAddress" bson:"shippingAddress"`
	ShippingMethod   string     `json:"shippingMethod" bson:"shippingMethod"`
	ShippingProvider string     `json:"shippingProvider" bson:"shippingProvider"`
	TrackingNumber   string     `json:"trackingNumber" bson:"trackingNumber"`
	ShippedDate      *time.Time `json:"shippedDate" bson:"shippedDate"`
	DeliveredDate    *time.Time `json:"deliveredDate" bson:"deliveredDate"`

	Notes         string `json:"notes" bson:"notes"`
	InternalNotes string `json:"internalNotes" bson:"internalNotes"`
	Terms         string `json:"terms" bson:"terms"`

	Discounts    []OrderDiscount    `json:"discounts" bson:"discounts"`
	Taxes        []OrderTax         `json:"taxes" bson:"taxes"`
	Payments     []OrderPayment     `json:"payments" bson:"payments"`
	Refunds      []OrderRefund      `json:"refunds" bson:"refunds"`
	Fulfillments []OrderFulfillment `json:"fulfillments" bson:"fulfillments"`

	Metadata map[string]string `json:"metadata" bson:"metadata"`
	Tags     []string          `json:"tags" bson:"tags"`

	QuoteID   *uuid.UUID `json:"quoteId" bson:"quoteId"`
	InvoiceID *uuid.UUID `json:"invoiceId" bson:"invoiceId"`

	Channel string `json:"channel" bson:"channel"`
	Locale  string `json:"locale" bson:"locale"`

	CreatedBy   uuid.UUID  `json:"createdBy" bson:"createdBy"`
	UpdatedBy   uuid.UUID  `json:"updatedBy" bson:"updatedBy"`
	ConfirmedBy *uuid.UUID `json:"confirmedBy" bson:"confirmedBy"`
	ConfirmedAt *time.Time `json:"confirmedAt" bson:"confirmedAt"`
	CreatedAt   time.Time  `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt" bson:"updatedAt"`
	Version     int64      `json:"-" bson:"version"`
}

type OrderLine struct {
	ID            uuid.UUID              `json:"id" bson:"_id"`
	ProductID     uuid.UUID              `json:"productId" bson:"productId"`
	VariantID     *uuid.UUID             `json:"variantId" bson:"variantId"`
	SKU           string                 `json:"sku" bson:"sku"`
	Name          string                 `json:"name" bson:"name"`
	Description   string                 `json:"description" bson:"description"`
	Quantity      int                    `json:"quantity" bson:"quantity"`
	UnitPrice     decimal.Decimal        `json:"unitPrice" bson:"unitPrice"`
	UnitCost      decimal.Decimal        `json:"unitCost" bson:"unitCost"`
	Discount      decimal.Decimal        `json:"discount" bson:"discount"`
	DiscountType  string                 `json:"discountType" bson:"discountType"`
	TaxRate       decimal.Decimal        `json:"taxRate" bson:"taxRate"`
	TaxAmount     decimal.Decimal        `json:"taxAmount" bson:"taxAmount"`
	RowTotal      decimal.Decimal        `json:"rowTotal" bson:"rowTotal"`
	RowCost       decimal.Decimal        `json:"rowCost" bson:"rowCost"`
	Weight        decimal.Decimal        `json:"weight" bson:"weight"`
	WeightUnit    string                 `json:"weightUnit" bson:"weightUnit"`
	FulfilledQty  int                    `json:"fulfilledQty" bson:"fulfilledQty"`
	ReservedQty   int                    `json:"reservedQty" bson:"reservedQty"`
	ShippedQty    int                    `json:"shippedQty" bson:"shippedQty"`
	ReturnableQty int                    `json:"returnableQty" bson:"returnableQty"`
	Position      int                    `json:"position" bson:"position"`
	CustomFields  map[string]interface{} `json:"customFields" bson:"customFields"`
}

type OrderDiscount struct {
	ID          uuid.UUID       `json:"id" bson:"_id"`
	Type        string          `json:"type" bson:"type"`
	Code        string          `json:"code" bson:"code"`
	Description string          `json:"description" bson:"description"`
	Amount      decimal.Decimal `json:"amount" bson:"amount"`
}

type OrderTax struct {
	ID          uuid.UUID       `json:"id" bson:"_id"`
	Name        string          `json:"name" bson:"name"`
	Rate        decimal.Decimal `json:"rate" bson:"rate"`
	Amount      decimal.Decimal `json:"amount" bson:"amount"`
	TaxCategory string          `json:"taxCategory" bson:"taxCategory"`
}

type OrderPayment struct {
	ID            uuid.UUID          `json:"id" bson:"_id"`
	Method        string             `json:"method" bson:"method"`
	Amount        decimal.Decimal    `json:"amount" bson:"amount"`
	Currency      string             `json:"currency" bson:"currency"`
	TransactionID string             `json:"transactionId" bson:"transactionId"`
	Provider      string             `json:"provider" bson:"provider"`
	Status        OrderPaymentStatus `json:"status" bson:"status"`
	ProcessedAt   *time.Time         `json:"processedAt" bson:"processedAt"`
}

type OrderRefund struct {
	ID            uuid.UUID       `json:"id" bson:"_id"`
	Amount        decimal.Decimal `json:"amount" bson:"amount"`
	Reason        string          `json:"reason" bson:"reason"`
	Method        string          `json:"method" bson:"method"`
	TransactionID string          `json:"transactionId" bson:"transactionId"`
	ProcessedBy   uuid.UUID       `json:"processedBy" bson:"processedBy"`
	ProcessedAt   time.Time       `json:"processedAt" bson:"processedAt"`
}

type OrderFulfillment struct {
	ID             uuid.UUID         `json:"id" bson:"_id"`
	TrackingNumber string            `json:"trackingNumber" bson:"trackingNumber"`
	Carrier        string            `json:"carrier" bson:"carrier"`
	Method         string            `json:"method" bson:"method"`
	Status         FulfillmentStatus `json:"status" bson:"status"`
	ShippedDate    *time.Time        `json:"shippedDate" bson:"shippedDate"`
	DeliveredDate  *time.Time        `json:"deliveredDate" bson:"deliveredDate"`
	Lines          []FulfillmentLine `json:"lines" bson:"lines"`
}

type FulfillmentLine struct {
	OrderLineID uuid.UUID `json:"orderLineId" bson:"orderLineId"`
	Quantity    int       `json:"quantity" bson:"quantity"`
}

func NewOrder(
	tenantID, clientID, createdBy uuid.UUID,
	orderType OrderType,
	source OrderSource,
	currency string,
) (*Order, error) {
	now := time.Now().UTC()
	id := uuid.New()

	order := &Order{
		ID:                id,
		TenantID:          tenantID,
		ClientID:          clientID,
		Type:              orderType,
		Source:            source,
		Status:            OrderStatusDraft,
		PaymentStatus:     OrderPaymentStatusPending,
		FulfillmentStatus: FulfillmentStatusUnfulfilled,
		Currency:          currency,
		Lines:             []OrderLine{},
		Discounts:         []OrderDiscount{},
		Taxes:             []OrderTax{},
		Payments:          []OrderPayment{},
		Refunds:           []OrderRefund{},
		Fulfillments:      []OrderFulfillment{},
		Metadata:          make(map[string]string),
		Tags:              []string{},
		Subtotal:          decimal.Zero,
		DiscountTotal:     decimal.Zero,
		TaxTotal:          decimal.Zero,
		ShippingTotal:     decimal.Zero,
		HandlingTotal:     decimal.Zero,
		Total:             decimal.Zero,
		AmountPaid:        decimal.Zero,
		AmountDue:         decimal.Zero,
		CreatedBy:         createdBy,
		CreatedAt:         now,
		UpdatedAt:         now,
		Version:           0,
	}

	return order, nil
}

func (o *Order) AddLine(line OrderLine) {
	line.ID = uuid.New()
	line.Position = len(o.Lines) + 1
	line.RowTotal = line.UnitPrice.Mul(decimal.NewFromInt(int64(line.Quantity))).Sub(line.Discount)
	line.RowCost = line.UnitCost.Mul(decimal.NewFromInt(int64(line.Quantity)))
	o.Lines = append(o.Lines, line)
	o.recalculate()
}

func (o *Order) RemoveLine(lineID uuid.UUID) {
	newLines := make([]OrderLine, 0, len(o.Lines))
	for _, line := range o.Lines {
		if line.ID != lineID {
			newLines = append(newLines, line)
		}
	}
	o.Lines = newLines
	o.recalculate()
}

func (o *Order) recalculate() {
	subtotal := decimal.Zero
	for _, line := range o.Lines {
		subtotal = subtotal.Add(line.RowTotal)
	}

	o.Subtotal = subtotal
	o.Total = o.Subtotal.Add(o.TaxTotal).Add(o.ShippingTotal).Add(o.HandlingTotal).Sub(o.DiscountTotal)
	o.AmountDue = o.Total.Sub(o.AmountPaid)
	o.UpdatedAt = time.Now().UTC()
}

func (o *Order) SetBillingAddress(addr *Address) {
	o.BillingAddress = addr
	o.UpdatedAt = time.Now().UTC()
}

func (o *Order) SetShippingAddress(addr *Address) {
	o.ShippingAddress = addr
	o.UpdatedAt = time.Now().UTC()
}

func (o *Order) SetShippingMethod(method, provider string, cost decimal.Decimal) {
	o.ShippingMethod = method
	o.ShippingProvider = provider
	o.ShippingTotal = cost
	o.recalculate()
}

func (o *Order) Confirm() {
	if o.Status == OrderStatusDraft || o.Status == OrderStatusPending {
		o.Status = OrderStatusConfirmed
		now := time.Now().UTC()
		o.ConfirmedAt = &now
		o.UpdatedAt = now
	}
}

func (o *Order) Process() {
	if o.Status == OrderStatusConfirmed {
		o.Status = OrderStatusProcessing
		o.UpdatedAt = time.Now().UTC()
	}
}

func (o *Order) Ship(trackingNumber, carrier string) {
	o.Status = OrderStatusShipped
	o.TrackingNumber = trackingNumber
	o.ShippingProvider = carrier
	now := time.Now().UTC()
	o.ShippedDate = &now
	o.FulfillmentStatus = FulfillmentStatusFulfilled
	o.UpdatedAt = now
}

func (o *Order) Deliver() {
	o.Status = OrderStatusDelivered
	now := time.Now().UTC()
	o.DeliveredDate = &now
	o.UpdatedAt = now
}

func (o *Order) Complete() {
	if o.Status == OrderStatusDelivered || o.Status == OrderStatusShipped {
		o.Status = OrderStatusCompleted
		o.UpdatedAt = time.Now().UTC()
	}
}

func (o *Order) Cancel(reason string) {
	o.Status = OrderStatusCancelled
	o.InternalNotes = o.InternalNotes + "\nCancelled: " + reason
	o.UpdatedAt = time.Now().UTC()
}

func (o *Order) AddPayment(payment OrderPayment) {
	payment.ID = uuid.New()
	o.Payments = append(o.Payments, payment)
	o.AmountPaid = o.AmountPaid.Add(payment.Amount)
	o.recalculate()

	if o.AmountPaid.GreaterThanOrEqual(o.Total) {
		o.PaymentStatus = OrderPaymentStatusPaid
	} else if o.AmountPaid.GreaterThan(decimal.Zero) {
		o.PaymentStatus = OrderPaymentStatusPartially
	}
}

func (o *Order) AddRefund(refund OrderRefund) {
	refund.ID = uuid.New()
	o.Refunds = append(o.Refunds, refund)

	totalRefunded := decimal.Zero
	for _, r := range o.Refunds {
		totalRefunded = totalRefunded.Add(r.Amount)
	}

	if totalRefunded.GreaterThanOrEqual(o.Total) {
		o.PaymentStatus = OrderPaymentStatusRefunded
		o.Status = OrderStatusRefunded
	} else {
		o.PaymentStatus = OrderPaymentStatusPartially
	}
	o.UpdatedAt = time.Now().UTC()
}

func (o *Order) ApplyDiscount(discount OrderDiscount) {
	discount.ID = uuid.New()
	o.Discounts = append(o.Discounts, discount)
	o.DiscountTotal = o.DiscountTotal.Add(discount.Amount)
	o.recalculate()
}

func (o *Order) AddTax(tax OrderTax) {
	tax.ID = uuid.New()
	o.Taxes = append(o.Taxes, tax)
	o.TaxTotal = o.TaxTotal.Add(tax.Amount)
	o.recalculate()
}

func (o *Order) GetTotalWeight() decimal.Decimal {
	total := decimal.Zero
	for _, line := range o.Lines {
		total = total.Add(line.Weight.Mul(decimal.NewFromInt(int64(line.Quantity))))
	}
	return total
}

func (o *Order) GetItemCount() int {
	count := 0
	for _, line := range o.Lines {
		count += line.Quantity
	}
	return count
}

func (o *Order) IsFullyPaid() bool {
	return o.AmountPaid.GreaterThanOrEqual(o.Total)
}

func (o *Order) IsFullyFulfilled() bool {
	fulfilledQty := 0
	totalQty := 0
	for _, line := range o.Lines {
		totalQty += line.Quantity
		fulfilledQty += line.FulfilledQty
	}
	return totalQty > 0 && fulfilledQty >= totalQty
}

func (o *Order) GetStatus() OrderStatus {
	return o.Status
}

var ErrOrderNotEditable = &OrderError{
	Code:    "ORDER_NOT_EDITABLE",
	Message: "Order cannot be edited in current status",
}

type OrderError struct {
	Code    string
	Message string
}

func (e *OrderError) Error() string {
	return e.Message
}
