import { provideHttpClient } from '@angular/common/http';
import {
  HttpTestingController,
  provideHttpClientTesting
} from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { Invoice } from './invoice.model';
import { InvoiceService } from './invoice.service';

describe('InvoiceService', () => {
  let service: InvoiceService;
  let http: HttpTestingController;

  const invoices: Invoice[] = [
    { id: 1, number: 'NF-0001', status: 'OPEN' },
    { id: 2, number: 'NF-0002', status: 'CLOSED' }
  ];

  const load = (list: Invoice[] = invoices) => {
    service.list().subscribe();
    http.expectOne('/api/billing/invoices').flush(list);
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()]
    });

    service = TestBed.inject(InvoiceService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  describe('list', () => {
    it('starts empty until the first load', () => {
      expect(service.invoices()).toEqual([]);
    });

    it('fetches the listing from the billing endpoint', () => {
      service.list().subscribe();

      const request = http.expectOne('/api/billing/invoices');
      expect(request.request.method).toBe('GET');

      request.flush(invoices);
    });

    it('publishes to the signal whatever the API returned', () => {
      load();
      expect(service.invoices()).toEqual(invoices);
    });

    it('replaces the previous list on every load', () => {
      load();
      load([invoices[0]]);

      expect(service.invoices()).toEqual([invoices[0]]);
    });

    it('propagates the failure and keeps the previous list', () => {
      load();

      let failed = false;
      service.list().subscribe({ error: () => (failed = true) });
      http
        .expectOne('/api/billing/invoices')
        .flush(null, { status: 500, statusText: 'Error' });

      expect(failed).toBe(true);
      expect(service.invoices()).toEqual(invoices);
    });
  });
});
