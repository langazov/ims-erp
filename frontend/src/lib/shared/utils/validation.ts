import { z } from 'zod';

// Client Validation Schema
export const clientSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters'),
  email: z.string().min(1, 'Email is required').email('Invalid email format'),
  phone: z.string().optional(),
  creditLimit: z.number().min(0, 'Credit limit must be positive').optional(),
  billingAddress: z.object({
    street: z.string().min(1, 'Street is required'),
    city: z.string().min(1, 'City is required'),
    state: z.string().min(1, 'State is required'),
    postalCode: z.string().min(1, 'Postal code is required'),
    country: z.string().min(1, 'Country is required')
  }).optional()
});

export type ClientFormData = z.infer<typeof clientSchema>;

// Product Validation Schema
export const productSchema = z.object({
  sku: z.string().min(1, 'SKU is required').max(50, 'SKU must be less than 50 characters'),
  name: z.string().min(1, 'Name is required').max(200, 'Name must be less than 200 characters'),
  description: z.string().max(2000, 'Description must be less than 2000 characters').optional(),
  category: z.string().min(1, 'Category is required'),
  type: z.enum(['physical', 'digital', 'service']),
  status: z.enum(['active', 'inactive', 'draft']),
  listPrice: z.number().min(0, 'Price must be positive'),
  salePrice: z.number().min(0, 'Sale price must be positive').optional(),
  costPrice: z.number().min(0, 'Cost price must be positive').optional(),
  stockQuantity: z.number().int().min(0, 'Stock quantity must be positive').optional(),
  lowStockThreshold: z.number().int().min(0, 'Threshold must be positive').optional()
});

export type ProductFormData = z.infer<typeof productSchema>;

// User Validation Schema
export const userSchema = z.object({
  email: z.string().min(1, 'Email is required').email('Invalid email format'),
  firstName: z.string().min(1, 'First name is required').max(50, 'First name must be less than 50 characters'),
  lastName: z.string().min(1, 'Last name is required').max(50, 'Last name must be less than 50 characters'),
  role: z.enum(['admin', 'manager', 'user', 'viewer']),
  department: z.string().max(100, 'Department must be less than 100 characters').optional(),
  phone: z.string().optional()
});

export type UserFormData = z.infer<typeof userSchema>;

// Warehouse Validation Schema
export const warehouseSchema = z.object({
  code: z.string().min(1, 'Code is required').max(20, 'Code must be less than 20 characters'),
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters'),
  type: z.enum(['main', 'distribution', 'retail', 'virtual']),
  address: z.object({
    street: z.string().min(1, 'Street is required'),
    city: z.string().min(1, 'City is required'),
    state: z.string().min(1, 'State is required'),
    postalCode: z.string().min(1, 'Postal code is required'),
    country: z.string().min(1, 'Country is required')
  }),
  capacity: z.number().int().min(0, 'Capacity must be positive').optional()
});

export type WarehouseFormData = z.infer<typeof warehouseSchema>;

// Invoice Validation Schema
export const invoiceSchema = z.object({
  clientId: z.string().min(1, 'Client is required'),
  issueDate: z.string().min(1, 'Issue date is required'),
  dueDate: z.string().min(1, 'Due date is required'),
  lineItems: z.array(z.object({
    description: z.string().min(1, 'Description is required'),
    quantity: z.number().int().min(1, 'Quantity must be at least 1'),
    unitPrice: z.number().min(0, 'Unit price must be positive')
  })).min(1, 'At least one line item is required'),
  notes: z.string().max(2000, 'Notes must be less than 2000 characters').optional()
});

export type InvoiceFormData = z.infer<typeof invoiceSchema>;

// Payment Validation Schema
export const paymentSchema = z.object({
  invoiceId: z.string().min(1, 'Invoice is required'),
  amount: z.number().min(0.01, 'Amount must be greater than 0'),
  method: z.enum(['credit_card', 'bank_transfer', 'cash', 'check', 'paypal', 'stripe']),
  reference: z.string().max(100, 'Reference must be less than 100 characters').optional(),
  notes: z.string().max(1000, 'Notes must be less than 1000 characters').optional()
});

export type PaymentFormData = z.infer<typeof paymentSchema>;

// Order Validation Schema
export const orderSchema = z.object({
  clientId: z.string().min(1, 'Client is required'),
  orderDate: z.string().min(1, 'Order date is required'),
  items: z.array(z.object({
    productId: z.string().min(1, 'Product is required'),
    quantity: z.number().int().min(1, 'Quantity must be at least 1'),
    unitPrice: z.number().min(0, 'Unit price must be positive')
  })).min(1, 'At least one item is required'),
  shippingAddress: z.object({
    street: z.string().min(1, 'Street is required'),
    city: z.string().min(1, 'City is required'),
    state: z.string().min(1, 'State is required'),
    postalCode: z.string().min(1, 'Postal code is required'),
    country: z.string().min(1, 'Country is required')
  })
});

export type OrderFormData = z.infer<typeof orderSchema>;

// Document Validation Schema
export const documentSchema = z.object({
  name: z.string().min(1, 'Name is required').max(200, 'Name must be less than 200 characters'),
  type: z.enum(['invoice', 'po', 'receipt', 'contract', 'scanned', 'other']),
  tags: z.array(z.string()).max(10, 'Maximum 10 tags allowed').optional()
});

export type DocumentFormData = z.infer<typeof documentSchema>;

// Validation helper function
export function validateForm<T>(
  schema: z.ZodSchema<T>,
  data: unknown
): { success: true; data: T } | { success: false; errors: Record<string, string> } {
  const result = schema.safeParse(data);
  
  if (result.success) {
    return { success: true, data: result.data };
  }
  
  const errors: Record<string, string> = {};
  result.error.errors.forEach((error) => {
    const path = error.path.join('.');
    errors[path] = error.message;
  });
  
  return { success: false, errors };
}

// Field-level validation helper
export function validateField<T>(
  schema: z.ZodSchema<T>,
  field: string,
  value: unknown
): string | undefined {
  const fieldSchema = schema instanceof z.ZodObject 
    ? schema.shape[field] 
    : schema;
    
  if (!fieldSchema) return undefined;
  
  const result = fieldSchema.safeParse(value);
  
  if (!result.success) {
    return result.error.errors[0]?.message;
  }
  
  return undefined;
}
