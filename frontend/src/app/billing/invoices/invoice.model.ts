export type InvoiceStatus = 'OPEN' | 'CLOSING' | 'CLOSED';
export type InvoiceType = 'IN' | 'OUT';

export interface InvoiceItem {
  id: number;
  productId: number;
  inventoryId: number;
  code: string;
  name: string;
  unit: string;
  quantity: number;
  unitPrice: number;
  total: number;
}

export interface Invoice {
  id: number;
  number: string;
  type: InvoiceType;
  status: InvoiceStatus;
  items: InvoiceItem[];
  total: number;
  failureReason?: string;
}

export interface InvoiceItemPayload {
  inventoryId: number;
  code: string;
  name: string;
  unit: string;
  quantity: number;
  unitPrice: number;
}

export interface InvoicePayload {
  number: string;
  type?: InvoiceType;
  items: InvoiceItemPayload[];
}

export const INVOICE_TYPE_LABELS: Record<InvoiceType, string> = {
  OUT: 'Saída',
  IN: 'Entrada'
};

export const INVOICE_TYPES = Object.keys(INVOICE_TYPE_LABELS) as InvoiceType[];

export const INVOICE_STATUS_LABELS: Record<InvoiceStatus, string> = {
  OPEN: 'Aberta',
  CLOSING: 'Processando',
  CLOSED: 'Fechada'
};

export const INVOICE_FAILURE_LABELS: Record<string, string> = {
  INSUFFICIENT_STOCK: 'Estoque insuficiente',
  PRODUCT_NOT_FOUND: 'Produto não encontrado no estoque'
};

