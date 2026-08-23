export type InvoiceStatus = 'OPEN' | 'CLOSING' | 'CLOSED' | 'REOPENING';
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

export interface InvoiceShortage {
  inventoryId: number;
  code: string;
  name: string;
  required: number;
  available: number;
}

export interface Invoice {
  id: number;
  series: number;
  number: number;
  formattedNumber: string;
  type: InvoiceType;
  status: InvoiceStatus;
  items: InvoiceItem[];
  total: number;
  failureReason?: string;
  shortages?: InvoiceShortage[];
  processingSince?: string;
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
  series: number;
  number: number;
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
  REOPENING: 'Processando',
  CLOSED: 'Fechada'
};

export const INVOICE_FAILURE_LABELS: Record<string, string> = {
  INSUFFICIENT_STOCK: 'Estoque insuficiente',
  PRODUCT_NOT_FOUND: 'Produto não encontrado no estoque',
  STOCK_ALREADY_USED: 'O estoque desta nota já foi utilizado'
};

export function isProcessing(invoice: Invoice): boolean {
  return invoice.status === 'CLOSING' || invoice.status === 'REOPENING';
}

export type ProcessingStage = 'normal' | 'unstable' | 'stuck';

export const UNSTABLE_AFTER = 30_000;
export const STUCK_AFTER = 5 * 60_000;

export const PROCESSING_STAGE_MESSAGES: Record<ProcessingStage, string> = {
  normal: '',
  unstable:
    'Estamos com instabilidade. A nota fiscal continua sendo processada, você não precisa fazer nada.',
  stuck: 'Não foi possível concluir o processamento desta nota fiscal.'
};

export function processingStage(invoice: Invoice, now: number): ProcessingStage {
  if (!isProcessing(invoice) || !invoice.processingSince) {
    return 'normal';
  }

  const elapsed = now - Date.parse(invoice.processingSince);

  if (elapsed >= STUCK_AFTER) {
    return 'stuck';
  }

  if (elapsed >= UNSTABLE_AFTER) {
    return 'unstable';
  }

  return 'normal';
}

