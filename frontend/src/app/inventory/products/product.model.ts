export interface Product {
  id: number;
  code: string;
  name: string;
  unit: string;
  price: number;
  stock: number;
}

export type ProductPayload = Omit<Product, 'id' | 'stock'> & { stock?: number };
