import { HttpClient } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { Observable, tap } from 'rxjs';

import { Product } from './product.model';

const RESOURCE = '/api/inventory/products';

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

  create(data: Omit<Product, 'id'>): Observable<Product> {
    return this.http
      .post<Product>(RESOURCE, data)
      .pipe(tap((product) => this._products.update((products) => [...products, product])));
  }
}
