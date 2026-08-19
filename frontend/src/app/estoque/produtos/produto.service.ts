import { Injectable, signal } from '@angular/core';
import { Produto } from './produto.model';

/**
 * Dados em memória enquanto a API não está disponível.
 */
const PRODUTOS_MOCK: Produto[] = [
  {
    id: 1,
    codigo: 'PRD-0001',
    descricao: 'Notebook Dell Inspiron 15',
    unidade: 'UN',
    precoUnitario: 4299.9,
    estoque: 12,
    ativo: true
  },
  {
    id: 2,
    codigo: 'PRD-0002',
    descricao: 'Monitor LG 24" Full HD',
    unidade: 'UN',
    precoUnitario: 899.0,
    estoque: 34,
    ativo: true
  },
  {
    id: 3,
    codigo: 'PRD-0003',
    descricao: 'Teclado Mecânico ABNT2',
    unidade: 'UN',
    precoUnitario: 349.5,
    estoque: 0,
    ativo: true
  },
  {
    id: 4,
    codigo: 'PRD-0004',
    descricao: 'Cabo HDMI 2.0 - 2 metros',
    unidade: 'PC',
    precoUnitario: 39.9,
    estoque: 128,
    ativo: true
  },
  {
    id: 5,
    codigo: 'PRD-0005',
    descricao: 'Papel Sulfite A4 75g (resma)',
    unidade: 'CX',
    precoUnitario: 27.4,
    estoque: 56,
    ativo: false
  }
];

@Injectable({ providedIn: 'root' })
export class ProdutoService {
  private readonly _produtos = signal<Produto[]>(PRODUTOS_MOCK);

  readonly produtos = this._produtos.asReadonly();

  /** Verifica se o código já está em uso, ignorando caixa e espaços. */
  existeCodigo(codigo: string): boolean {
    const alvo = codigo.trim().toUpperCase();
    return this._produtos().some((produto) => produto.codigo.toUpperCase() === alvo);
  }

  /** Sugere o próximo código sequencial no padrão PRD-0000. */
  proximoCodigo(): string {
    const ultimo = this._produtos().reduce((maior, produto) => {
      const numero = Number(/^PRD-(\d+)$/.exec(produto.codigo)?.[1]);
      return Number.isFinite(numero) && numero > maior ? numero : maior;
    }, 0);

    return `PRD-${String(ultimo + 1).padStart(4, '0')}`;
  }

  cadastrar(dados: Omit<Produto, 'id'>): Produto {
    const id = this._produtos().reduce((maior, produto) => Math.max(maior, produto.id), 0) + 1;
    const produto: Produto = { ...dados, id, codigo: dados.codigo.trim().toUpperCase() };

    this._produtos.update((produtos) => [...produtos, produto]);

    return produto;
  }
}
