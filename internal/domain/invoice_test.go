package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInvoice(t *testing.T) {
	tenantID := uuid.New()
	clientID := uuid.New()
	createdBy := uuid.New()

	invoice, err := NewInvoice(tenantID, clientID, createdBy, InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	require.NoError(t, err)
	assert.NotEmpty(t, invoice.ID)
	assert.Equal(t, tenantID, invoice.TenantID)
	assert.Equal(t, clientID, invoice.ClientID)
	assert.Equal(t, InvoiceTypeStandard, invoice.Type)
	assert.Equal(t, InvoiceStatusDraft, invoice.Status)
	assert.Equal(t, "USD", invoice.Currency)
	assert.Equal(t, PaymentTermNet30, invoice.PaymentTerm)
	assert.True(t, invoice.Subtotal.IsZero())
	assert.True(t, invoice.TaxTotal.IsZero())
	assert.True(t, invoice.Total.IsZero())
	assert.True(t, invoice.AmountPaid.IsZero())
	assert.True(t, invoice.AmountDue.IsZero())
	assert.Empty(t, invoice.Lines)
	assert.False(t, invoice.CreatedAt.IsZero())
}

func TestInvoiceAddLine(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	line := InvoiceLine{
		Description: "Test Product",
		Quantity:    decimal.NewFromInt(2),
		UnitPrice:   decimal.NewFromFloat(50.00),
		Discount:    decimal.Zero,
		TaxRate:     decimal.NewFromFloat(10),
	}
	invoice.AddLine(line)

	require.Len(t, invoice.Lines, 1)
	assert.Equal(t, "Test Product", invoice.Lines[0].Description)
	assert.True(t, invoice.Subtotal.Equal(decimal.NewFromFloat(100.00)))
}

func TestInvoiceRemoveLine(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	line := InvoiceLine{
		Description: "Test Product",
		Quantity:    decimal.NewFromInt(1),
		UnitPrice:   decimal.NewFromFloat(100.00),
		Discount:    decimal.Zero,
		TaxRate:     decimal.NewFromFloat(0),
	}
	invoice.AddLine(line)
	lineID := invoice.Lines[0].ID

	invoice.RemoveLine(lineID)

	assert.Empty(t, invoice.Lines)
	assert.True(t, invoice.Subtotal.IsZero())
}

func TestInvoiceRecalculate(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	line1 := InvoiceLine{
		Description: "Product A",
		Quantity:    decimal.NewFromInt(2),
		UnitPrice:   decimal.NewFromFloat(50.00),
		Discount:    decimal.NewFromFloat(10.00),
		TaxRate:     decimal.NewFromFloat(10),
	}
	line2 := InvoiceLine{
		Description: "Product B",
		Quantity:    decimal.NewFromInt(1),
		UnitPrice:   decimal.NewFromFloat(100.00),
		Discount:    decimal.Zero,
		TaxRate:     decimal.NewFromFloat(10),
	}

	invoice.AddLine(line1)
	invoice.AddLine(line2)

	assert.False(t, invoice.Subtotal.IsZero())
	assert.False(t, invoice.TaxTotal.IsZero())
	assert.False(t, invoice.Total.IsZero())
	assert.True(t, invoice.Total.Equal(invoice.Subtotal.Add(invoice.TaxTotal).Sub(invoice.DiscountTotal)))
}

func TestInvoiceSend(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	assert.Equal(t, InvoiceStatusDraft, invoice.Status)

	invoice.Send()

	assert.Equal(t, InvoiceStatusSent, invoice.Status)
	assert.NotNil(t, invoice.SentDate)
}

func TestInvoiceMarkAsSent(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	invoice.MarkAsSent()

	assert.Equal(t, InvoiceStatusSent, invoice.Status)
}

func TestInvoiceSetStatus(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	invoice.SetStatus(InvoiceStatusPending)

	assert.Equal(t, InvoiceStatusPending, invoice.Status)
}

func TestInvoiceMarkAsPaid(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	line := InvoiceLine{
		Description: "Product",
		Quantity:    decimal.NewFromInt(1),
		UnitPrice:   decimal.NewFromFloat(100.00),
		Discount:    decimal.Zero,
		TaxRate:     decimal.NewFromFloat(0),
	}
	invoice.AddLine(line)

	invoice.MarkAsPaid(decimal.NewFromFloat(50.00))

	assert.True(t, invoice.AmountPaid.Equal(decimal.NewFromFloat(50.00)))
	assert.True(t, invoice.AmountDue.Equal(decimal.NewFromFloat(50.00)))

	invoice.MarkAsPaid(decimal.NewFromFloat(50.00))

	assert.True(t, invoice.AmountPaid.Equal(invoice.Total))
	assert.True(t, invoice.AmountDue.IsZero())
	assert.Equal(t, InvoiceStatusPaid, invoice.Status)
	assert.NotNil(t, invoice.PaidDate)
}

func TestInvoiceApplyPayment(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	line := InvoiceLine{
		Description: "Product",
		Quantity:    decimal.NewFromInt(1),
		UnitPrice:   decimal.NewFromFloat(100.00),
		Discount:    decimal.Zero,
		TaxRate:     decimal.NewFromFloat(0),
	}
	invoice.AddLine(line)

	err := invoice.ApplyPayment(decimal.NewFromFloat(30.00))
	require.NoError(t, err)
	assert.True(t, invoice.AmountPaid.Equal(decimal.NewFromFloat(30.00)))

	err = invoice.ApplyPayment(decimal.NewFromFloat(100.00))
	assert.Error(t, err)
}

func TestInvoiceCancel(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	invoice.Cancel("Duplicate invoice")

	assert.Equal(t, InvoiceStatusCancelled, invoice.Status)
	assert.Contains(t, invoice.Notes, "Duplicate invoice")
}

func TestInvoiceRefund(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())
	invoice.SetStatus(InvoiceStatusPaid)

	invoice.Refund()

	assert.Equal(t, InvoiceStatusRefunded, invoice.Status)
}

func TestInvoiceSetDueDate(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	newDueDate := time.Now().AddDate(0, 0, 15)
	invoice.SetDueDate(newDueDate)

	assert.NotNil(t, invoice.DueDate)
	assert.Equal(t, newDueDate.Day(), invoice.DueDate.Day())
}

func TestInvoiceSetNotes(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	invoice.SetNotes("Thank you for your business")

	assert.Equal(t, "Thank you for your business", invoice.Notes)
}

func TestInvoiceSetTerms(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	invoice.SetTerms("Payment due within 30 days")

	assert.Equal(t, "Payment due within 30 days", invoice.Terms)
}

func TestInvoiceSetInvoiceNumber(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	invoice.SetInvoiceNumber("INV-2024-001")

	assert.Equal(t, "INV-2024-001", invoice.InvoiceNumber)
}

func TestInvoiceCalculateDueDate(t *testing.T) {
	issueDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, issueDate)

	dueDate := invoice.CalculateDueDate()

	assert.Equal(t, issueDate.AddDate(0, 0, 30).Day(), dueDate.Day())
}

func TestInvoiceIsOverdue(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	pastDueDate := time.Now().AddDate(0, 0, -5)
	invoice.DueDate = &pastDueDate
	invoice.Send()

	assert.True(t, invoice.IsOverdue())

	futureDueDate := time.Now().AddDate(0, 0, 5)
	invoice.DueDate = &futureDueDate

	assert.False(t, invoice.IsOverdue())
}

func TestInvoiceGetDaysOverdue(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	assert.Equal(t, 0, invoice.GetDaysOverdue())
}

func TestInvoiceGetPaymentProgress(t *testing.T) {
	invoice, _ := NewInvoice(uuid.New(), uuid.New(), uuid.New(), InvoiceTypeStandard, "USD", PaymentTermNet30, time.Now())

	assert.True(t, invoice.GetPaymentProgress().IsZero())

	line := InvoiceLine{
		Description: "Product",
		Quantity:    decimal.NewFromInt(1),
		UnitPrice:   decimal.NewFromFloat(100.00),
		Discount:    decimal.Zero,
		TaxRate:     decimal.NewFromFloat(0),
	}
	invoice.AddLine(line)

	invoice.MarkAsPaid(decimal.NewFromFloat(25.00))

	progress := invoice.GetPaymentProgress()
	assert.True(t, progress.GreaterThan(decimal.NewFromFloat(24.9)))
	assert.True(t, progress.LessThan(decimal.NewFromFloat(25.1)))
}

func TestInvoiceLineCalculations(t *testing.T) {
	line := InvoiceLine{
		Description: "Product",
		Quantity:    decimal.NewFromInt(3),
		UnitPrice:   decimal.NewFromFloat(25.00),
		Discount:    decimal.NewFromFloat(10.00),
		TaxRate:     decimal.NewFromFloat(8),
	}

	expectedTotal := decimal.NewFromFloat(65.00)
	expectedTaxAmount := decimal.NewFromFloat(5.20)

	line.Total = line.Quantity.Mul(line.UnitPrice).Sub(line.Discount)
	line.TaxAmount = line.Total.Mul(line.TaxRate).Div(decimal.NewFromInt(100))

	assert.True(t, line.Total.Equal(expectedTotal))
	assert.True(t, line.TaxAmount.Equal(expectedTaxAmount))
}
