import { HttpClient } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { Observable, tap } from 'rxjs';

import { Invoice, InvoicePayload } from './invoice.model';

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

  create(data: InvoicePayload): Observable<Invoice> {
    return this.http
      .post<Invoice>(RESOURCE, data)
      .pipe(tap((invoice) => this._invoices.update((invoices) => [...invoices, invoice])));
  }

  update(id: number, data: InvoicePayload): Observable<Invoice> {
    return this.http.put<Invoice>(`${RESOURCE}/${id}`, data).pipe(
      tap((updated) =>
        this._invoices.update((invoices) =>
          invoices.map((invoice) => (invoice.id === updated.id ? updated : invoice))
        )
      )
    );
  }

  remove(id: number): Observable<void> {
    return this.http
      .delete<void>(`${RESOURCE}/${id}`)
      .pipe(
        tap(() =>
          this._invoices.update((invoices) => invoices.filter((invoice) => invoice.id !== id))
        )
      );
  }

  danfeUrl(id: number): string {
    return `${RESOURCE}/${id}/danfe`;
  }

  close(id: number): Observable<Invoice> {
    return this.transition(id, 'close');
  }

  reopen(id: number): Observable<Invoice> {
    return this.transition(id, 'reopen');
  }

  private transition(id: number, action: 'close' | 'reopen'): Observable<Invoice> {
    return this.http.post<Invoice>(`${RESOURCE}/${id}/${action}`, {}).pipe(
      tap((moved) =>
        this._invoices.update((invoices) =>
          invoices.map((invoice) => (invoice.id === moved.id ? moved : invoice))
        )
      )
    );
  }
}
