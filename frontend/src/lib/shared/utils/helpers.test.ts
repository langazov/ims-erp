import { describe, it, expect } from 'vitest';
import {
  cn,
  formatDate,
  formatDateTime,
  formatCurrency,
  formatNumber,
  formatPercentage,
  formatPhoneNumber,
  truncate,
  generateId,
  debounce,
  throttle,
  getInitials,
  slugify,
  capitalize,
  pluralize,
  isValidEmail,
  isValidPhone,
  groupBy,
  sortBy,
  uniqueBy
} from './helpers';

describe('cn (className utility)', () => {
  it('should merge class names correctly', () => {
    expect(cn('class1', 'class2')).toBe('class1 class2');
    expect(cn('class1', false && 'class2', 'class3')).toBe('class1 class3');
    expect(cn('px-2', 'px-4')).toBe('px-4'); // Tailwind merge
  });
});

describe('formatDate', () => {
  it('should format date correctly', () => {
    const date = new Date('2024-01-15');
    expect(formatDate(date)).toBe('Jan 15, 2024');
    expect(formatDate('2024-01-15')).toBe('Jan 15, 2024');
  });

  it('should accept custom options', () => {
    const date = new Date('2024-01-15');
    expect(formatDate(date, { month: 'long' })).toBe('January 15, 2024');
  });
});

describe('formatDateTime', () => {
  it('should format date and time correctly', () => {
    const date = new Date('2024-01-15T14:30:00');
    const result = formatDateTime(date);
    expect(result).toContain('Jan 15, 2024');
    expect(result).toContain(':');
  });
});

describe('formatCurrency', () => {
  it('should format currency correctly', () => {
    expect(formatCurrency(1000)).toBe('$1,000.00');
    expect(formatCurrency(1000.50)).toBe('$1,000.50');
    expect(formatCurrency('500')).toBe('$500.00');
  });

  it('should handle different currencies', () => {
    expect(formatCurrency(1000, 'EUR')).toContain('€');
    expect(formatCurrency(1000, 'GBP')).toContain('£');
  });
});

describe('formatNumber', () => {
  it('should format numbers with commas', () => {
    expect(formatNumber(1000)).toBe('1,000');
    expect(formatNumber(1000000)).toBe('1,000,000');
  });
});

describe('formatPercentage', () => {
  it('should format percentages correctly', () => {
    expect(formatPercentage(50)).toBe('50%');
    expect(formatPercentage(50.5, 1)).toBe('50.5%');
    expect(formatPercentage(50.555, 2)).toBe('50.55%');
  });
});

describe('formatPhoneNumber', () => {
  it('should format US phone numbers', () => {
    expect(formatPhoneNumber('1234567890')).toBe('(123) 456-7890');
    expect(formatPhoneNumber('123-456-7890')).toBe('(123) 456-7890');
  });

  it('should return original if invalid', () => {
    expect(formatPhoneNumber('123')).toBe('123');
  });
});

describe('truncate', () => {
  it('should truncate long strings', () => {
    expect(truncate('Hello World', 5)).toBe('Hello...');
    expect(truncate('Hi', 10)).toBe('Hi');
  });
});

describe('generateId', () => {
  it('should generate unique IDs', () => {
    const id1 = generateId();
    const id2 = generateId();
    expect(id1).not.toBe(id2);
    expect(id1.length).toBe(7);
  });
});

describe('debounce', () => {
  it('should debounce function calls', async () => {
    let count = 0;
    const fn = debounce(() => count++, 100);
    
    fn();
    fn();
    fn();
    
    expect(count).toBe(0);
    await new Promise(r => setTimeout(r, 150));
    expect(count).toBe(1);
  });
});

describe('throttle', () => {
  it('should throttle function calls', async () => {
    let count = 0;
    const fn = throttle(() => count++, 100);
    
    fn();
    fn();
    fn();
    
    expect(count).toBe(1);
    await new Promise(r => setTimeout(r, 150));
    fn();
    expect(count).toBe(2);
  });
});

describe('getInitials', () => {
  it('should get initials from name', () => {
    expect(getInitials('John Doe')).toBe('JD');
    expect(getInitials('Jane')).toBe('J');
    expect(getInitials('john doe')).toBe('JD');
  });
});

describe('slugify', () => {
  it('should convert text to slug', () => {
    expect(slugify('Hello World')).toBe('hello-world');
    expect(slugify('Test & Example')).toBe('test-example');
    expect(slugify('Multiple   Spaces')).toBe('multiple-spaces');
  });
});

describe('capitalize', () => {
  it('should capitalize first letter', () => {
    expect(capitalize('hello')).toBe('Hello');
    expect(capitalize('Hello')).toBe('Hello');
  });
});

describe('pluralize', () => {
  it('should pluralize correctly', () => {
    expect(pluralize(1, 'item')).toBe('item');
    expect(pluralize(2, 'item')).toBe('items');
    expect(pluralize(2, 'child', 'children')).toBe('children');
  });
});

describe('isValidEmail', () => {
  it('should validate email addresses', () => {
    expect(isValidEmail('test@example.com')).toBe(true);
    expect(isValidEmail('invalid')).toBe(false);
    expect(isValidEmail('test@')).toBe(false);
  });
});

describe('isValidPhone', () => {
  it('should validate phone numbers', () => {
    expect(isValidPhone('1234567890')).toBe(true);
    expect(isValidPhone('123-456-7890')).toBe(true);
    expect(isValidPhone('123')).toBe(false);
  });
});

describe('groupBy', () => {
  it('should group array by key', () => {
    const items = [
      { category: 'A', name: 'Item 1' },
      { category: 'B', name: 'Item 2' },
      { category: 'A', name: 'Item 3' }
    ];
    
    const grouped = groupBy(items, 'category');
    expect(grouped.A).toHaveLength(2);
    expect(grouped.B).toHaveLength(1);
  });
});

describe('sortBy', () => {
  it('should sort array by key', () => {
    const items = [
      { name: 'Charlie', age: 30 },
      { name: 'Alice', age: 25 },
      { name: 'Bob', age: 35 }
    ];
    
    const sorted = sortBy(items, 'name');
    expect(sorted[0].name).toBe('Alice');
    expect(sorted[1].name).toBe('Bob');
    expect(sorted[2].name).toBe('Charlie');
  });

  it('should handle descending order', () => {
    const items = [{ value: 1 }, { value: 3 }, { value: 2 }];
    const sorted = sortBy(items, 'value', 'desc');
    expect(sorted[0].value).toBe(3);
    expect(sorted[2].value).toBe(1);
  });
});

describe('uniqueBy', () => {
  it('should filter unique items by key', () => {
    const items = [
      { id: 1, name: 'A' },
      { id: 2, name: 'B' },
      { id: 1, name: 'C' }
    ];
    
    const unique = uniqueBy(items, 'id');
    expect(unique).toHaveLength(2);
  });
});
