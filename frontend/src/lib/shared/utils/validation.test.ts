import { describe, it, expect } from 'vitest';
import { z } from 'zod';
import {
  clientSchema,
  productSchema,
  userSchema,
  warehouseSchema,
  invoiceSchema,
  paymentSchema,
  orderSchema,
  documentSchema,
  validateForm,
  validateField
} from './validation';

describe('Client Schema', () => {
  it('should validate valid client data', () => {
    const validClient = {
      name: 'John Doe',
      email: 'john@example.com',
      phone: '123-456-7890'
    };
    
    const result = validateForm(clientSchema, validClient);
    expect(result.success).toBe(true);
  });

  it('should reject invalid email', () => {
    const invalidClient = {
      name: 'John Doe',
      email: 'invalid-email'
    };
    
    const result = validateForm(clientSchema, invalidClient);
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.errors.email).toBeDefined();
    }
  });

  it('should require name', () => {
    const invalidClient = {
      name: '',
      email: 'john@example.com'
    };
    
    const result = validateForm(clientSchema, invalidClient);
    expect(result.success).toBe(false);
  });
});

describe('Product Schema', () => {
  it('should validate valid product data', () => {
    const validProduct = {
      sku: 'PROD-001',
      name: 'Test Product',
      category: 'Electronics',
      type: 'physical',
      status: 'active',
      listPrice: 99.99
    };
    
    const result = validateForm(productSchema, validProduct);
    expect(result.success).toBe(true);
  });

  it('should reject negative price', () => {
    const invalidProduct = {
      sku: 'PROD-001',
      name: 'Test Product',
      category: 'Electronics',
      type: 'physical',
      status: 'active',
      listPrice: -10
    };
    
    const result = validateForm(productSchema, invalidProduct);
    expect(result.success).toBe(false);
  });
});

describe('User Schema', () => {
  it('should validate valid user data', () => {
    const validUser = {
      email: 'user@example.com',
      firstName: 'John',
      lastName: 'Doe',
      role: 'user'
    };
    
    const result = validateForm(userSchema, validUser);
    expect(result.success).toBe(true);
  });

  it('should reject invalid role', () => {
    const invalidUser = {
      email: 'user@example.com',
      firstName: 'John',
      lastName: 'Doe',
      role: 'invalid-role'
    };
    
    const result = validateForm(userSchema, invalidUser);
    expect(result.success).toBe(false);
  });
});

describe('Warehouse Schema', () => {
  it('should validate valid warehouse data', () => {
    const validWarehouse = {
      code: 'WH001',
      name: 'Main Warehouse',
      type: 'main',
      address: {
        street: '123 Main St',
        city: 'New York',
        state: 'NY',
        postalCode: '10001',
        country: 'USA'
      }
    };
    
    const result = validateForm(warehouseSchema, validWarehouse);
    expect(result.success).toBe(true);
  });

  it('should require complete address', () => {
    const invalidWarehouse = {
      code: 'WH001',
      name: 'Main Warehouse',
      type: 'main',
      address: {
        street: '123 Main St',
        city: '',
        state: 'NY',
        postalCode: '10001',
        country: 'USA'
      }
    };
    
    const result = validateForm(warehouseSchema, invalidWarehouse);
    expect(result.success).toBe(false);
  });
});

describe('Invoice Schema', () => {
  it('should validate valid invoice data', () => {
    const validInvoice = {
      clientId: '1',
      issueDate: '2024-01-15',
      dueDate: '2024-02-15',
      lineItems: [
        { description: 'Service', quantity: 1, unitPrice: 100 }
      ]
    };
    
    const result = validateForm(invoiceSchema, validInvoice);
    expect(result.success).toBe(true);
  });

  it('should require at least one line item', () => {
    const invalidInvoice = {
      clientId: '1',
      issueDate: '2024-01-15',
      dueDate: '2024-02-15',
      lineItems: []
    };
    
    const result = validateForm(invoiceSchema, invalidInvoice);
    expect(result.success).toBe(false);
  });
});

describe('Payment Schema', () => {
  it('should validate valid payment data', () => {
    const validPayment = {
      invoiceId: '1',
      amount: 100,
      method: 'credit_card'
    };
    
    const result = validateForm(paymentSchema, validPayment);
    expect(result.success).toBe(true);
  });

  it('should reject zero amount', () => {
    const invalidPayment = {
      invoiceId: '1',
      amount: 0,
      method: 'credit_card'
    };
    
    const result = validateForm(paymentSchema, invalidPayment);
    expect(result.success).toBe(false);
  });
});

describe('Order Schema', () => {
  it('should validate valid order data', () => {
    const validOrder = {
      clientId: '1',
      orderDate: '2024-01-15',
      items: [
        { productId: '1', quantity: 2, unitPrice: 50 }
      ],
      shippingAddress: {
        street: '123 Main St',
        city: 'New York',
        state: 'NY',
        postalCode: '10001',
        country: 'USA'
      }
    };
    
    const result = validateForm(orderSchema, validOrder);
    expect(result.success).toBe(true);
  });
});

describe('Document Schema', () => {
  it('should validate valid document data', () => {
    const validDocument = {
      name: 'Invoice_001.pdf',
      type: 'invoice'
    };
    
    const result = validateForm(documentSchema, validDocument);
    expect(result.success).toBe(true);
  });
});

describe('validateForm helper', () => {
  it('should return success true for valid data', () => {
    const schema = z.object({ name: z.string() });
    const result = validateForm(schema, { name: 'Test' });
    
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.name).toBe('Test');
    }
  });

  it('should return errors for invalid data', () => {
    const schema = z.object({ 
      name: z.string().min(1, 'Name is required'),
      email: z.string().email('Invalid email')
    });
    const result = validateForm(schema, { name: '', email: 'invalid' });
    
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(Object.keys(result.errors).length).toBeGreaterThan(0);
    }
  });
});

describe('validateField helper', () => {
  it('should validate individual fields', () => {
    const schema = z.object({ 
      email: z.string().email('Invalid email')
    });
    
    const error = validateField(schema, 'email', 'invalid');
    expect(error).toBe('Invalid email');
    
    const noError = validateField(schema, 'email', 'test@example.com');
    expect(noError).toBeUndefined();
  });
});
