package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusPending   InvoiceStatus = "pending"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
	InvoiceStatusRefunded  InvoiceStatus = "refunded"
)

type InvoiceType string

const (
	InvoiceTypeStandard   InvoiceType = "standard"
	InvoiceTypeCreditNote InvoiceType = "credit_note"
	InvoiceTypeDebitNote  InvoiceType = "debit_note"
	InvoiceTypeRecurring  InvoiceType = "recurring"
)

type PaymentTerm string

const (
	PaymentTermDueOnReceipt PaymentTerm = "due_on_receipt"
	PaymentTermNet15        PaymentTerm = "net_15"
	PaymentTermNet30        PaymentTerm = "net_30"
	PaymentTermNet45        PaymentTerm = "net_45"
	PaymentTermNet60        PaymentTerm = "net_60"
	PaymentTermEndOfMonth   PaymentTerm = "end_of_month"
)

type Invoice struct {
	ID            uuid.UUID         `json:"id" bson:"_id"`
	TenantID      uuid.UUID         `json:"tenantId" bson:"tenantId"`
	InvoiceNumber string            `json:"invoiceNumber" bson:"invoiceNumber"`
	ClientID      uuid.UUID         `json:"clientId" bson:"clientId"`
	Type          InvoiceType       `json:"type" bson:"type"`
	Status        InvoiceStatus     `json:"status" bson:"status"`
	Currency      string            `json:"currency" bson:"currency"`
	Subtotal      decimal.Decimal   `json:"subtotal" bson:"subtotal"`
	TaxTotal      decimal.Decimal   `json:"taxTotal" bson:"taxTotal"`
	DiscountTotal decimal.Decimal   `json:"discountTotal" bson:"discountTotal"`
	Total         decimal.Decimal   `json:"total" bson:"total"`
	AmountPaid    decimal.Decimal   `json:"amountPaid" bson:"amountPaid"`
	AmountDue     decimal.Decimal   `json:"amountDue" bson:"amountDue"`
	PaymentTerm   PaymentTerm       `json:"paymentTerm" bson:"paymentTerm"`
	DueDate       *time.Time        `json:"dueDate" bson:"dueDate"`
	IssueDate     time.Time         `json:"issueDate" bson:"issueDate"`
	SentDate      *time.Time        `json:"sentDate" bson:"sentDate"`
	PaidDate      *time.Time        `json:"paidDate" bson:"paidDate"`
	Lines         []InvoiceLine     `json:"lines" bson:"lines"`
	Notes         string            `json:"notes" bson:"notes"`
	Terms         string            `json:"terms" bson:"terms"`
	AttachmentURL string            `json:"attachmentUrl" bson:"attachmentUrl"`
	Metadata      map[string]string `json:"metadata" bson:"metadata"`
	CreatedBy     uuid.UUID         `json:"createdBy" bson:"createdBy"`
	CreatedAt     time.Time         `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt" bson:"updatedAt"`
	Version       int64             `json:"-" bson:"version"`
}

type InvoiceLine struct {
	ID          uuid.UUID       `json:"id" bson:"id"`
	Description string          `json:"description" bson:"description"`
	Quantity    decimal.Decimal `json:"quantity" bson:"quantity"`
	UnitPrice   decimal.Decimal `json:"unitPrice" bson:"unitPrice"`
	Discount    decimal.Decimal `json:"discount" bson:"discount"`
	TaxRate     decimal.Decimal `json:"taxRate" bson:"taxRate"`
	TaxAmount   decimal.Decimal `json:"taxAmount" bson:"taxAmount"`
	Total       decimal.Decimal `json:"total" bson:"total"`
	ProductID   *uuid.UUID      `json:"productId" bson:"productId"`
	ServiceDate *time.Time      `json:"serviceDate" bson:"serviceDate"`
	SortOrder   int             `json:"sortOrder" bson:"sortOrder"`
}

func NewInvoice(
	tenantID, clientID, createdBy uuid.UUID,
	invoiceType InvoiceType,
	currency string,
	paymentTerm PaymentTerm,
	issueDate time.Time,
) (*Invoice, error) {
	now := time.Now().UTC()
	id := uuid.New()

	invoice := &Invoice{
		ID:            id,
		TenantID:      tenantID,
		ClientID:      clientID,
		Type:          invoiceType,
		Status:        InvoiceStatusDraft,
		Currency:      currency,
		PaymentTerm:   paymentTerm,
		IssueDate:     issueDate,
		Lines:         []InvoiceLine{},
		Subtotal:      decimal.Zero,
		TaxTotal:      decimal.Zero,
		DiscountTotal: decimal.Zero,
		Total:         decimal.Zero,
		AmountPaid:    decimal.Zero,
		AmountDue:     decimal.Zero,
		CreatedBy:     createdBy,
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       0,
	}

	return invoice, nil
}

func (i *Invoice) AddLine(line InvoiceLine) {
	line.Total = line.Quantity.Mul(line.UnitPrice).Sub(line.Discount)
	line.TaxAmount = line.Total.Mul(line.TaxRate).Div(decimal.NewFromInt(100))
	line.ID = uuid.New()
	i.Lines = append(i.Lines, line)
	i.recalculate()
}

func (i *Invoice) RemoveLine(lineID uuid.UUID) {
	newLines := make([]InvoiceLine, 0, len(i.Lines))
	for _, line := range i.Lines {
		if line.ID != lineID {
			newLines = append(newLines, line)
		}
	}
	i.Lines = newLines
	i.recalculate()
}

func (i *Invoice) UpdateLine(lineID uuid.UUID, update func(*InvoiceLine)) {
	for idx, line := range i.Lines {
		if line.ID == lineID {
			update(&i.Lines[idx])
			break
		}
	}
	i.recalculate()
}

func (i *Invoice) recalculate() {
	subtotal := decimal.Zero
	taxTotal := decimal.Zero
	discountTotal := decimal.Zero

	for _, line := range i.Lines {
		subtotal = subtotal.Add(line.Total)
		taxTotal = taxTotal.Add(line.TaxAmount)
		discountTotal = discountTotal.Add(line.Discount)
	}

	i.Subtotal = subtotal
	i.TaxTotal = taxTotal
	i.DiscountTotal = discountTotal
	i.Total = subtotal.Add(taxTotal).Sub(discountTotal)
	i.AmountDue = i.Total.Sub(i.AmountPaid)
	i.UpdatedAt = time.Now().UTC()
}

func (i *Invoice) SetDueDate(dueDate time.Time) {
	i.DueDate = &dueDate
	i.UpdatedAt = time.Now().UTC()
}

func (i *Invoice) SetInvoiceNumber(number string) {
	i.InvoiceNumber = number
	i.UpdatedAt = time.Now().UTC()
}

func (i *Invoice) SetNotes(notes string) {
	i.Notes = notes
	i.UpdatedAt = time.Now().UTC()
}

func (i *Invoice) SetTerms(terms string) {
	i.Terms = terms
	i.UpdatedAt = time.Now().UTC()
}

func (i *Invoice) Send() {
	if i.Status == InvoiceStatusDraft || i.Status == InvoiceStatusPending {
		now := time.Now().UTC()
		i.Status = InvoiceStatusSent
		i.SentDate = &now
		i.UpdatedAt = now
	}
}

func (i *Invoice) MarkAsSent() {
	i.Send()
}

func (i *Invoice) SetStatus(status InvoiceStatus) {
	i.Status = status
	i.UpdatedAt = time.Now().UTC()
}

func (i *Invoice) MarkAsPaid(amount decimal.Decimal) {
	i.AmountPaid = i.AmountPaid.Add(amount)
	i.AmountDue = i.Total.Sub(i.AmountPaid)

	if i.AmountDue.LessThanOrEqual(decimal.Zero) {
		i.Status = InvoiceStatusPaid
		now := time.Now().UTC()
		i.PaidDate = &now
	}
	i.UpdatedAt = time.Now().UTC()
}

func (i *Invoice) ApplyPayment(amount decimal.Decimal) error {
	if amount.GreaterThan(i.AmountDue) {
		return ErrPaymentExceedsAmount
	}
	i.MarkAsPaid(amount)
	return nil
}

func (i *Invoice) Cancel(reason string) {
	i.Status = InvoiceStatusCancelled
	i.Notes = i.Notes + "\nCancelled: " + reason
	i.UpdatedAt = time.Now().UTC()
}

func (i *Invoice) Refund() {
	if i.Status == InvoiceStatusPaid {
		i.Status = InvoiceStatusRefunded
		i.UpdatedAt = time.Now().UTC()
	}
}

func (i *Invoice) CalculateDueDate() time.Time {
	var days int
	switch i.PaymentTerm {
	case PaymentTermDueOnReceipt:
		days = 0
	case PaymentTermNet15:
		days = 15
	case PaymentTermNet30:
		days = 30
	case PaymentTermNet45:
		days = 45
	case PaymentTermNet60:
		days = 60
	case PaymentTermEndOfMonth:
		return endOfMonth(i.IssueDate)
	default:
		days = 30
	}
	return i.IssueDate.AddDate(0, 0, days)
}

func endOfMonth(t time.Time) time.Time {
	lastDay := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location())
	return lastDay
}

func (i *Invoice) IsOverdue() bool {
	if i.DueDate == nil {
		return false
	}
	return i.Status == InvoiceStatusSent && time.Now().UTC().After(*i.DueDate)
}

func (i *Invoice) GetDaysOverdue() int {
	if i.DueDate == nil {
		return 0
	}
	if !i.IsOverdue() {
		return 0
	}
	return int(time.Now().UTC().Sub(*i.DueDate).Hours() / 24)
}

func (i *Invoice) GetPaymentProgress() decimal.Decimal {
	if i.Total.IsZero() {
		return decimal.Zero
	}
	return i.AmountPaid.Div(i.Total).Mul(decimal.NewFromInt(100))
}

var ErrPaymentExceedsAmount = &PaymentError{
	Code:    "PAYMENT_EXCEEDS_AMOUNT",
	Message: "Payment amount exceeds the amount due",
}
