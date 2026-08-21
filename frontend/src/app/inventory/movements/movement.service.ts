import { HttpClient } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { Observable, tap } from 'rxjs';

import { Movement, MovementPayload } from './movement.model';

const RESOURCE = '/api/inventory/products';

@Injectable({ providedIn: 'root' })
export class MovementService {
  private readonly http = inject(HttpClient);

  private readonly _movements = signal<Movement[]>([]);

  readonly movements = this._movements.asReadonly();

  list(productId: number): Observable<Movement[]> {
    return this.http
      .get<Movement[]>(this.resource(productId))
      .pipe(tap((movements) => this._movements.set(movements)));
  }

  get(productId: number, id: number): Observable<Movement> {
    return this.http.get<Movement>(`${this.resource(productId)}/${id}`);
  }

  create(productId: number, data: MovementPayload): Observable<Movement> {
    return this.http
      .post<Movement>(this.resource(productId), data)
      .pipe(tap((movement) => this._movements.update((movements) => [movement, ...movements])));
  }

  update(productId: number, id: number, data: MovementPayload): Observable<Movement> {
    return this.http.put<Movement>(`${this.resource(productId)}/${id}`, data).pipe(
      tap((updated) =>
        this._movements.update((movements) =>
          movements.map((movement) => (movement.id === updated.id ? updated : movement))
        )
      )
    );
  }

  private resource(productId: number): string {
    return `${RESOURCE}/${productId}/movements`;
  }
}
