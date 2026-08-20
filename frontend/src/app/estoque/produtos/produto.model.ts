export interface Produto {
  id: number;
  code: string;
  name: string;
  unit: string;
  price: number;
  stock: number;
}

export type ErrosDoServidor = Record<string, string>;
