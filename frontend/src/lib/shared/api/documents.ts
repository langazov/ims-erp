// Documents API
import api from './index';

export type DocumentType = 'invoice' | 'contract' | 'receipt' | 'report' | 'image' | 'other';
export type DocumentStatus = 'uploaded' | 'processing' | 'processed' | 'failed';

export interface Document {
  id: string;
  tenantId: string;
  name: string;
  originalName: string;
  type: DocumentType;
  status: DocumentStatus;
  mimeType: string;
  size: number;
  url: string;
  thumbnailUrl: string | null;
  ocrText: string | null;
  entityType: string | null;
  entityId: string | null;
  tags: string[];
  uploadedBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface DocumentFilter {
  type?: DocumentType;
  status?: DocumentStatus;
  entityType?: string;
  entityId?: string;
  search?: string;
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface DocumentListResponse {
  data: Document[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface UploadDocumentRequest {
  name?: string;
  type?: DocumentType;
  entityType?: string;
  entityId?: string;
  tags?: string[];
}

const BASE_PATH = '/documents';

export async function getDocuments(filter?: DocumentFilter): Promise<DocumentListResponse> {
  const params = filter as Record<string, string | number | boolean | undefined>;
  return api.get(`${BASE_PATH}`, params);
}

export async function getDocumentById(id: string): Promise<Document> {
  return api.get(`${BASE_PATH}/${id}`);
}

export async function uploadDocument(file: File, metadata?: UploadDocumentRequest): Promise<Document> {
  const formData = new FormData();
  formData.append('file', file);
  if (metadata) {
    Object.entries(metadata).forEach(([key, value]) => {
      if (value !== undefined) {
        formData.append(key, Array.isArray(value) ? value.join(',') : String(value));
      }
    });
  }
  const API_BASE = import.meta.env.VITE_API_URL || '/api';
  const response = await fetch(`${API_BASE}/documents`, {
    method: 'POST',
    body: formData,
    credentials: 'include',
  });
  if (!response.ok) {
    const err = await response.json().catch(() => ({ message: 'Upload failed' }));
    throw new Error(err.message);
  }
  return response.json();
}

export async function updateDocument(id: string, data: Partial<UploadDocumentRequest>): Promise<Document> {
  return api.patch(`${BASE_PATH}/${id}`, data);
}

export async function deleteDocument(id: string): Promise<void> {
  return api.delete(`${BASE_PATH}/${id}`);
}

export async function searchDocuments(query: string, filter?: Omit<DocumentFilter, 'search'>): Promise<DocumentListResponse> {
  return api.get(`${BASE_PATH}/search`, { q: query, ...filter });
}

export async function getDocumentUrl(id: string): Promise<{ url: string; expiresAt: string }> {
  return api.get(`${BASE_PATH}/${id}/url`);
}

export async function linkDocumentToEntity(id: string, entityType: string, entityId: string): Promise<Document> {
  return api.post(`${BASE_PATH}/${id}/link`, { entityType, entityId });
}

export async function getOcrText(id: string): Promise<{ text: string; confidence: number }> {
  return api.get(`${BASE_PATH}/${id}/ocr`);
}
