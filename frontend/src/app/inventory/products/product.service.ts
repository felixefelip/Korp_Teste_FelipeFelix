import { HttpClient } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { Observable, tap } from 'rxjs';

import { Product, ProductPayload } from './product.model';

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

  get(id: number): Observable<Product> {
    return this.http.get<Product>(`${RESOURCE}/${id}`);
  }

  create(data: ProductPayload): Observable<Product> {
    return this.http
      .post<Product>(RESOURCE, data)
      .pipe(tap((product) => this._products.update((products) => [...products, product])));
  }

  update(id: number, data: ProductPayload): Observable<Product> {
    return this.http.put<Product>(`${RESOURCE}/${id}`, data).pipe(
      tap((updated) =>
        this._products.update((products) =>
          products.map((product) => (product.id === updated.id ? updated : product))
        )
      )
    );
  }
}
