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

  const newInvoice: Omit<Invoice, 'id'> = { number: 'NF-0100', status: 'OPEN' };

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

  describe('create', () => {
    it('posts the invoice and returns what the API created', () => {
      const created = { ...newInvoice, id: 7 };
      let received: Invoice | undefined;

      service.create(newInvoice).subscribe((invoice) => (received = invoice));

      const request = http.expectOne('/api/billing/invoices');
      expect(request.request.method).toBe('POST');
      expect(request.request.body).toEqual(newInvoice);

      request.flush(created);

      expect(received).toEqual(created);
    });

    it('appends to the listing the invoice returned by the API', () => {
      load();

      service.create(newInvoice).subscribe();
      http
        .expectOne('/api/billing/invoices')
        .flush({ ...newInvoice, id: 7, number: 'NF-0100' });

      expect(service.invoices()).toHaveLength(3);
      expect(service.invoices().at(-1)).toMatchObject({ id: 7, number: 'NF-0100' });
    });

    it('does not mutate the previous list array', () => {
      load();
      const previousList = service.invoices();

      service.create(newInvoice).subscribe();
      http.expectOne('/api/billing/invoices').flush({ ...newInvoice, id: 7 });

      expect(service.invoices()).not.toBe(previousList);
      expect(previousList).toHaveLength(2);
    });

    it('appends nothing when the API rejects the invoice', () => {
      load();

      let failed = false;
      service.create(newInvoice).subscribe({ error: () => (failed = true) });
      http
        .expectOne('/api/billing/invoices')
        .flush(
          { errors: { number: 'Campo obrigatório.' } },
          { status: 400, statusText: 'Bad Request' }
        );

      expect(failed).toBe(true);
      expect(service.invoices()).toEqual(invoices);
    });
  });
});
