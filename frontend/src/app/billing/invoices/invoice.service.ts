import { HttpClient } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { Observable, tap } from 'rxjs';

import { Invoice } from './invoice.model';

const RESOURCE = '/api/billing/invoices';

@Injectable({ providedIn: 'root' })
export class InvoiceService {
  private readonly http = inject(HttpClient);

  private readonly _invoices = signal<Invoice[]>([]);

  readonly invoices = this._invoices.asReadonly();

  list(): Observable<Invoice[]> {
    return this.http
      .get<Invoice[]>(RESOURCE)
      .pipe(tap((invoices) => this._invoices.set(invoices)));
  }

  get(id: number): Observable<Invoice> {
    return this.http.get<Invoice>(`${RESOURCE}/${id}`);
  }

  create(data: Omit<Invoice, 'id'>): Observable<Invoice> {
    return this.http
      .post<Invoice>(RESOURCE, data)
      .pipe(tap((invoice) => this._invoices.update((invoices) => [...invoices, invoice])));
  }

  update(id: number, data: Omit<Invoice, 'id'>): Observable<Invoice> {
    return this.http.put<Invoice>(`${RESOURCE}/${id}`, data).pipe(
      tap((updated) =>
        this._invoices.update((invoices) =>
          invoices.map((invoice) => (invoice.id === updated.id ? updated : invoice))
        )
      )
    );
  }
}
