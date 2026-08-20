export type InvoiceStatus = 'OPEN' | 'CLOSED';

export interface Invoice {
  id: number;
  number: string;
  status: InvoiceStatus;
}
