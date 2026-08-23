import { HttpClient } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { Observable, tap } from 'rxjs';

import { CatalogProduct } from './catalog.model';

const RESOURCE = '/api/billing/products';

@Injectable({ providedIn: 'root' })
export class CatalogService {
  private readonly http = inject(HttpClient);

  private readonly _products = signal<CatalogProduct[]>([]);

  readonly products = this._products.asReadonly();

  list(): Observable<CatalogProduct[]> {
    return this.http
      .get<CatalogProduct[]>(RESOURCE)
      .pipe(tap((products) => this._products.set(products)));
  }
}
