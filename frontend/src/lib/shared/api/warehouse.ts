// Warehouse & Location API
import api from './index';

export type WarehouseType = 'main' | 'distribution' | 'retail' | 'virtual';
export type WarehouseStatus = 'active' | 'inactive';
export type OperationType = 'receipt' | 'issue' | 'transfer' | 'adjustment' | 'count';
export type OperationStatus = 'pending' | 'in_progress' | 'completed' | 'cancelled';

export interface Warehouse {
  id: string;
  tenantId: string;
  code: string;
  name: string;
  type: WarehouseType;
  status: WarehouseStatus;
  address: {
    street: string;
    city: string;
    state: string;
    postalCode: string;
    country: string;
  };
  capacity: number;
  utilizedCapacity: number;
  locationCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface WarehouseLocation {
  id: string;
  warehouseId: string;
  code: string;
  name: string;
  zone: string;
  aisle: string;
  rack: string;
  bin: string;
  capacity: number;
  occupancy: number;
  isActive: boolean;
}

export interface WarehouseOperation {
  id: string;
  warehouseId: string;
  type: OperationType;
  status: OperationStatus;
  referenceType: string;
  referenceId: string;
  items: OperationItem[];
  notes: string;
  startedAt: string | null;
  completedAt: string | null;
  createdBy: string;
  createdAt: string;
}

export interface OperationItem {
  id: string;
  productId: string;
  productName: string;
  sku: string;
  locationId: string;
  locationCode: string;
  expectedQty: number;
  actualQty: number;
  status: 'pending' | 'completed';
}

export interface WarehouseFilter {
  status?: WarehouseStatus;
  type?: WarehouseType;
  search?: string;
  page?: number;
  pageSize?: number;
}

export interface WarehouseListResponse {
  data: Warehouse[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface CreateWarehouseRequest {
  code: string;
  name: string;
  type: WarehouseType;
  address?: Warehouse['address'];
  capacity?: number;
}

const BASE_PATH = '/warehouses';

export async function getWarehouses(filter?: WarehouseFilter): Promise<WarehouseListResponse> {
  const params = filter as Record<string, string | number | boolean | undefined>;
  return api.get(`${BASE_PATH}`, params);
}

export async function getWarehouseById(id: string): Promise<Warehouse> {
  return api.get(`${BASE_PATH}/${id}`);
}

export async function createWarehouse(data: CreateWarehouseRequest): Promise<Warehouse> {
  return api.post(`${BASE_PATH}`, data);
}

export async function updateWarehouse(id: string, data: Partial<CreateWarehouseRequest>): Promise<Warehouse> {
  return api.patch(`${BASE_PATH}/${id}`, data);
}

export async function deleteWarehouse(id: string): Promise<void> {
  return api.delete(`${BASE_PATH}/${id}`);
}

export async function getLocations(warehouseId: string): Promise<WarehouseLocation[]> {
  return api.get(`${BASE_PATH}/${warehouseId}/locations`);
}

export async function createLocation(warehouseId: string, data: Partial<WarehouseLocation>): Promise<WarehouseLocation> {
  return api.post(`${BASE_PATH}/${warehouseId}/locations`, data);
}

export async function updateLocation(warehouseId: string, locationId: string, data: Partial<WarehouseLocation>): Promise<WarehouseLocation> {
  return api.patch(`${BASE_PATH}/${warehouseId}/locations/${locationId}`, data);
}

export async function deleteLocation(warehouseId: string, locationId: string): Promise<void> {
  return api.delete(`${BASE_PATH}/${warehouseId}/locations/${locationId}`);
}

export async function getOperations(warehouseId: string): Promise<WarehouseOperation[]> {
  return api.get(`${BASE_PATH}/${warehouseId}/operations`);
}

export async function createOperation(warehouseId: string, data: Partial<WarehouseOperation>): Promise<WarehouseOperation> {
  return api.post(`${BASE_PATH}/${warehouseId}/operations`, data);
}

export async function startOperation(warehouseId: string, operationId: string): Promise<WarehouseOperation> {
  return api.post(`${BASE_PATH}/${warehouseId}/operations/${operationId}/start`, {});
}

export async function completeOperation(warehouseId: string, operationId: string): Promise<WarehouseOperation> {
  return api.post(`${BASE_PATH}/${warehouseId}/operations/${operationId}/complete`, {});
}
