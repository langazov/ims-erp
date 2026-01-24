package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ClientStatus string

const (
	ClientStatusActive    ClientStatus = "active"
	ClientStatusInactive  ClientStatus = "inactive"
	ClientStatusSuspended ClientStatus = "suspended"
	ClientStatusMerged    ClientStatus = "merged"
)

func (s ClientStatus) IsValid() bool {
	switch s {
	case ClientStatusActive, ClientStatusInactive, ClientStatusSuspended, ClientStatusMerged:
		return true
	}
	return false
}

type Address struct {
	Street     string `json:"street" bson:"street"`
	City       string `json:"city" bson:"city"`
	State      string `json:"state" bson:"state"`
	PostalCode string `json:"postalCode" bson:"postalCode"`
	Country    string `json:"country" bson:"country"`
}

func (a Address) IsEmpty() bool {
	return a.Street == "" && a.City == "" && a.Country == ""
}

type Client struct {
	ID                uuid.UUID
	TenantID          uuid.UUID
	Code              string
	Name              string
	Email             string
	Phone             string
	Status            ClientStatus
	CreditLimit       decimal.Decimal
	CurrentBalance    decimal.Decimal
	BillingAddress    Address
	ShippingAddresses []Address
	Tags              []string
	CustomFields      map[string]interface{}
	Version           int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NewClient(tenantID uuid.UUID, name, email string) *Client {
	now := time.Now().UTC()
	return &Client{
		ID:                uuid.New(),
		TenantID:          tenantID,
		Code:              "",
		Name:              name,
		Email:             email,
		Phone:             "",
		Status:            ClientStatusActive,
		CreditLimit:       decimal.Zero,
		CurrentBalance:    decimal.Zero,
		BillingAddress:    Address{},
		ShippingAddresses: []Address{},
		Tags:              []string{},
		CustomFields:      make(map[string]interface{}),
		Version:           0,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func (c *Client) Update(name, email, phone string) {
	c.Name = name
	c.Email = email
	c.Phone = phone
	c.UpdatedAt = time.Now().UTC()
}

func (c *Client) SetBillingAddress(address Address) {
	c.BillingAddress = address
	c.UpdatedAt = time.Now().UTC()
}

func (c *Client) AddShippingAddress(address Address) {
	c.ShippingAddresses = append(c.ShippingAddresses, address)
	c.UpdatedAt = time.Now().UTC()
}

func (c *Client) RemoveShippingAddress(index int) {
	if index >= 0 && index < len(c.ShippingAddresses) {
		c.ShippingAddresses = append(c.ShippingAddresses[:index], c.ShippingAddresses[index+1:]...)
		c.UpdatedAt = time.Now().UTC()
	}
}

func (c *Client) SetCreditLimit(limit decimal.Decimal) {
	c.CreditLimit = limit
	c.UpdatedAt = time.Now().UTC()
}

func (c *Client) UpdateBalance(amount decimal.Decimal) {
	c.CurrentBalance = c.CurrentBalance.Add(amount)
	c.UpdatedAt = time.Now().UTC()
}

func (c *Client) AvailableCredit() decimal.Decimal {
	return c.CreditLimit.Sub(c.CurrentBalance)
}

func (c *Client) CanMakePurchase(amount decimal.Decimal) bool {
	return c.CurrentBalance.Add(amount).LessThanOrEqual(c.CreditLimit)
}

func (c *Client) AddTag(tag string) {
	for _, t := range c.Tags {
		if t == tag {
			return
		}
	}
	c.Tags = append(c.Tags, tag)
	c.UpdatedAt = time.Now().UTC()
}

func (c *Client) RemoveTag(tag string) {
	for i, t := range c.Tags {
		if t == tag {
			c.Tags = append(c.Tags[:i], c.Tags[i+1:]...)
			c.UpdatedAt = time.Now().UTC()
			return
		}
	}
}

func (c *Client) SetCustomField(key string, value interface{}) {
	if c.CustomFields == nil {
		c.CustomFields = make(map[string]interface{})
	}
	c.CustomFields[key] = value
	c.UpdatedAt = time.Now().UTC()
}

func (c *Client) GetCustomField(key string) (interface{}, bool) {
	v, ok := c.CustomFields[key]
	return v, ok
}

func (c *Client) Deactivate() {
	c.Status = ClientStatusInactive
	c.UpdatedAt = time.Now().UTC()
}

func (c *Client) Suspend() {
	c.Status = ClientStatusSuspended
	c.UpdatedAt = time.Now().UTC()
}

func (c *Client) Reactivate() {
	c.Status = ClientStatusActive
	c.UpdatedAt = time.Now().UTC()
}

func (c *Client) MergeInto(target *Client) {
	for _, addr := range c.ShippingAddresses {
		target.AddShippingAddress(addr)
	}
	for _, tag := range c.Tags {
		target.AddTag(tag)
	}
	for k, v := range c.CustomFields {
		target.SetCustomField(k, v)
	}
	target.UpdatedAt = time.Now().UTC()
}

func (c *Client) HasOverdueBalance() bool {
	return c.CurrentBalance.GreaterThan(decimal.Zero)
}

func (c *Client) CreditUtilization() decimal.Decimal {
	if c.CreditLimit.IsZero() {
		return decimal.Zero
	}
	return c.CurrentBalance.Div(c.CreditLimit).Mul(decimal.NewFromInt(100))
}

type ClientEvent struct {
	ClientID  uuid.UUID
	EventType string
	EventData map[string]interface{}
	UserID    string
	Timestamp time.Time
}

func (c *Client) CreateEvent(eventType, userID string, data map[string]interface{}) *ClientEvent {
	return &ClientEvent{
		ClientID:  c.ID,
		EventType: eventType,
		EventData: data,
		UserID:    userID,
		Timestamp: time.Now().UTC(),
	}
}

type ClientActivityLog struct {
	ClientID  uuid.UUID
	Action    string
	Details   string
	UserID    string
	IPAddress string
	UserAgent string
	Timestamp time.Time
}
