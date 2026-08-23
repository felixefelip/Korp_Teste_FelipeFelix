import { provideHttpClient } from '@angular/common/http';
import {
  HttpTestingController,
  provideHttpClientTesting
} from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { Movement } from './movement.model';
import { MovementService } from './movement.service';

const RESOURCE = '/api/inventory/products/7/movements';

const ENTRY: Movement = {
  id: 1,
  productId: 7,
  type: 'in',
  origin: 'adjustment',
  quantity: 10,
  confirmed: true,
  billingInvoiceItemId: null,
  billingInvoiceId: null
};

const EXIT: Movement = {
  id: 2,
  productId: 7,
  type: 'out',
  origin: 'invoice',
  quantity: 4,
  confirmed: false,
  billingInvoiceItemId: 33,
  billingInvoiceId: 42
};

describe('MovementService', () => {
  let service: MovementService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()]
    });

    service = TestBed.inject(MovementService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('starts with an empty ledger', () => {
    expect(service.movements()).toEqual([]);
  });

  describe('list', () => {
    it('asks for the movements nested under the product', () => {
      service.list(7).subscribe();

      const request = http.expectOne(RESOURCE);

      expect(request.request.method).toBe('GET');
      request.flush([ENTRY, EXIT]);
    });

    it('publishes what came back on the signal', () => {
      service.list(7).subscribe();
      http.expectOne(RESOURCE).flush([ENTRY, EXIT]);

      expect(service.movements()).toEqual([ENTRY, EXIT]);
    });

    it('replaces the ledger of the previous product', () => {
      service.list(7).subscribe();
      http.expectOne(RESOURCE).flush([ENTRY]);

      service.list(9).subscribe();
      http.expectOne('/api/inventory/products/9/movements').flush([]);

      expect(service.movements()).toEqual([]);
    });
  });

  describe('get', () => {
    it('asks for that single movement', () => {
      service.get(7, 1).subscribe();

      const request = http.expectOne(`${RESOURCE}/1`);

      expect(request.request.method).toBe('GET');
      request.flush(ENTRY);
    });

    it('leaves the listed ledger alone', () => {
      service.list(7).subscribe();
      http.expectOne(RESOURCE).flush([ENTRY]);

      service.get(7, 1).subscribe();
      http.expectOne(`${RESOURCE}/1`).flush({ ...ENTRY, quantity: 999 });

      expect(service.movements()).toEqual([ENTRY]);
    });
  });

  describe('create', () => {
    it('posts the payload to the product ledger', () => {
      service.create(7, { type: 'in', quantity: 10, confirmed: true }).subscribe();

      const request = http.expectOne(RESOURCE);

      expect(request.request.method).toBe('POST');
      expect(request.request.body).toEqual({
        type: 'in',
        quantity: 10,
        confirmed: true
      });
      request.flush(ENTRY);
    });

    it('puts the new movement at the top of the ledger', () => {
      service.list(7).subscribe();
      http.expectOne(RESOURCE).flush([EXIT]);

      service.create(7, { type: 'in', quantity: 10, confirmed: true }).subscribe();
      http.expectOne(RESOURCE).flush(ENTRY);

      expect(service.movements()).toEqual([ENTRY, EXIT]);
    });

    it('leaves the ledger alone when the API fails', () => {
      service.list(7).subscribe();
      http.expectOne(RESOURCE).flush([EXIT]);

      service
        .create(7, { type: 'in', quantity: 10, confirmed: true })
        .subscribe({ error: () => undefined });
      http.expectOne(RESOURCE).flush('', { status: 400, statusText: 'Bad Request' });

      expect(service.movements()).toEqual([EXIT]);
    });
  });

  describe('update', () => {
    it('puts the payload on that movement', () => {
      service.update(7, 1, { type: 'out', quantity: 3, confirmed: false }).subscribe();

      const request = http.expectOne(`${RESOURCE}/1`);

      expect(request.request.method).toBe('PUT');
      expect(request.request.body).toEqual({
        type: 'out',
        quantity: 3,
        confirmed: false
      });
      request.flush(ENTRY);
    });

    it('swaps the edited movement in place', () => {
      service.list(7).subscribe();
      http.expectOne(RESOURCE).flush([ENTRY, EXIT]);

      service.update(7, 1, { type: 'in', quantity: 3, confirmed: true }).subscribe();
      http.expectOne(`${RESOURCE}/1`).flush({ ...ENTRY, quantity: 3 });

      expect(service.movements()).toEqual([{ ...ENTRY, quantity: 3 }, EXIT]);
    });
  });
});
