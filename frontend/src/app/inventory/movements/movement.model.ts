export type MovementType = 'in' | 'out';
export type MovementOrigin = 'adjustment' | 'sale';

export interface Movement {
  id: number;
  productId: number;
  type: MovementType;
  origin: MovementOrigin;
  quantity: number;
  confirmed: boolean;
  invoiceItemId: number | null;
}

export interface MovementPayload {
  type: MovementType;
  quantity: number;
  confirmed: boolean;
}

export const MOVEMENT_TYPE_LABELS: Record<MovementType, string> = {
  in: 'Entrada',
  out: 'Saída'
};

export const MOVEMENT_ORIGIN_LABELS: Record<MovementOrigin, string> = {
  adjustment: 'Ajuste',
  sale: 'Venda'
};

export function isFromInvoice(movement: Movement): boolean {
  return movement.invoiceItemId !== null;
}
