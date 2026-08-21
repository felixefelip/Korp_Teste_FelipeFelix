export type InvoiceStatus = 'OPEN' | 'CLOSED';

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
  status: InvoiceStatus;
  items: InvoiceItem[];
  total: number;
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
  status: InvoiceStatus;
  items: InvoiceItemPayload[];
}

export const INVOICE_STATUS_LABELS: Record<InvoiceStatus, string> = {
  OPEN: 'Aberta',
  CLOSED: 'Fechada'
};

export const INVOICE_STATUSES = Object.keys(INVOICE_STATUS_LABELS) as InvoiceStatus[];
