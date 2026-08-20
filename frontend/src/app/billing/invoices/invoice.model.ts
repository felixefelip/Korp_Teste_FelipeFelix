export type InvoiceStatus = 'OPEN' | 'CLOSED';

export interface Invoice {
  id: number;
  number: string;
  status: InvoiceStatus;
}

export const INVOICE_STATUS_LABELS: Record<InvoiceStatus, string> = {
  OPEN: 'Aberta',
  CLOSED: 'Fechada'
};

export const INVOICE_STATUSES = Object.keys(INVOICE_STATUS_LABELS) as InvoiceStatus[];
