import { HttpClient } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { Observable, tap } from 'rxjs';

import { Produto } from './produto.model';

const RECURSO = '/api/products';

@Injectable({ providedIn: 'root' })
export class ProdutoService {
  private readonly http = inject(HttpClient);

  private readonly _produtos = signal<Produto[]>([]);

  readonly produtos = this._produtos.asReadonly();

  listar(): Observable<Produto[]> {
    return this.http
      .get<Produto[]>(RECURSO)
      .pipe(tap((produtos) => this._produtos.set(produtos)));
  }

  proximoCodigo(): string {
    const ultimo = this._produtos().reduce((maior, produto) => {
      const numero = Number(/^PRD-(\d+)$/.exec(produto.code)?.[1]);
      return Number.isFinite(numero) && numero > maior ? numero : maior;
    }, 0);

    return `PRD-${String(ultimo + 1).padStart(4, '0')}`;
  }

  cadastrar(dados: Omit<Produto, 'id'>): Observable<Produto> {
    return this.http
      .post<Produto>(RECURSO, dados)
      .pipe(
        tap((produto) => this._produtos.update((produtos) => [...produtos, produto]))
      );
  }
}
