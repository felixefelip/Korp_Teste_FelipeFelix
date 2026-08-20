export interface Product {
  id: number;
  code: string;
  name: string;
  unit: string;
  price: number;
  stock: number;
}

export type ServerErrors = Record<string, string>;
