import { HttpClient } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { Observable, tap } from 'rxjs';

import { Product } from './product.model';

const RESOURCE = '/api/products';

@Injectable({ providedIn: 'root' })
export class ProductService {
  private readonly http = inject(HttpClient);

  private readonly _products = signal<Product[]>([]);

  readonly products = this._products.asReadonly();

  list(): Observable<Product[]> {
    return this.http
      .get<Product[]>(RESOURCE)
      .pipe(tap((products) => this._products.set(products)));
  }

  nextCode(): string {
    const last = this._products().reduce((highest, product) => {
      const number = Number(/^PRD-(\d+)$/.exec(product.code)?.[1]);
      return Number.isFinite(number) && number > highest ? number : highest;
    }, 0);

    return `PRD-${String(last + 1).padStart(4, '0')}`;
  }

  create(data: Omit<Product, 'id'>): Observable<Product> {
    return this.http
      .post<Product>(RESOURCE, data)
      .pipe(tap((product) => this._products.update((products) => [...products, product])));
  }
}
