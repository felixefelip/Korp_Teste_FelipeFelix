import { provideHttpClient } from '@angular/common/http';
import {
  HttpTestingController,
  provideHttpClientTesting
} from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { Invoice, InvoicePayload } from './invoice.model';
import { InvoiceService } from './invoice.service';

describe('InvoiceService', () => {
  let service: InvoiceService;
  let http: HttpTestingController;

  const newInvoice: InvoicePayload = { number: 'NF-0100', status: 'OPEN', items: [] };

  const withItem: InvoicePayload = {
    number: 'NF-0101',
    status: 'OPEN',
    items: [
      {
        inventoryId: 3,
        code: 'PRD-0003',
        name: 'Cadeira',
        unit: 'UN',
        quantity: 2,
        unitPrice: 150.5
      }
    ]
  };

  const invoices: Invoice[] = [
    { id: 1, number: 'NF-0001', status: 'OPEN', items: [], total: 0 },
    { id: 2, number: 'NF-0002', status: 'CLOSED', items: [], total: 0 }
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

    it('posts the items exactly as they were handed over', () => {
      service.create(withItem).subscribe();

      const request = http.expectOne('/api/billing/invoices');
      expect(request.request.body).toEqual(withItem);

      request.flush({ ...withItem, id: 8, total: 301 });
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

  describe('get', () => {
    it('fetches the invoice of that id', () => {
      let received: Invoice | undefined;

      service.get(7).subscribe((invoice) => (received = invoice));

      const request = http.expectOne('/api/billing/invoices/7');
      expect(request.request.method).toBe('GET');

      request.flush({ id: 7, number: 'NF-0007', status: 'CLOSED', items: [], total: 0 });

      expect(received).toEqual({
        id: 7,
        number: 'NF-0007',
        status: 'CLOSED',
        items: [],
        total: 0
      });
    });

    it('returns the items the API attached to the invoice', () => {
      let received: Invoice | undefined;

      service.get(7).subscribe((invoice) => (received = invoice));

      http.expectOne('/api/billing/invoices/7').flush({
        id: 7,
        number: 'NF-0007',
        status: 'OPEN',
        total: 301,
        items: [
          {
            id: 1,
            productId: 4,
            inventoryId: 3,
            code: 'PRD-0003',
            name: 'Cadeira',
            unit: 'UN',
            quantity: 2,
            unitPrice: 150.5,
            total: 301
          }
        ]
      });

      expect(received?.items).toHaveLength(1);
      expect(received?.items[0]).toMatchObject({ code: 'PRD-0003', quantity: 2 });
      expect(received?.total).toBe(301);
    });

    it('propagates the failure of an invoice that does not exist', () => {
      let status: number | undefined;

      service.get(404).subscribe({ error: (response) => (status = response.status) });
      http
        .expectOne('/api/billing/invoices/404')
        .flush(null, { status: 404, statusText: 'Not Found' });

      expect(status).toBe(404);
    });

    it('leaves the listing untouched', () => {
      load();

      service.get(1).subscribe();
      http.expectOne('/api/billing/invoices/1').flush(invoices[0]);

      expect(service.invoices()).toEqual(invoices);
    });
  });

  describe('update', () => {
    it('puts the invoice and returns what the API saved', () => {
      const updated = { ...newInvoice, id: 1 };
      let received: Invoice | undefined;

      service.update(1, newInvoice).subscribe((invoice) => (received = invoice));

      const request = http.expectOne('/api/billing/invoices/1');
      expect(request.request.method).toBe('PUT');
      expect(request.request.body).toEqual(newInvoice);

      request.flush(updated);

      expect(received).toEqual(updated);
    });

    it('puts the items exactly as they were handed over', () => {
      service.update(1, withItem).subscribe();

      const request = http.expectOne('/api/billing/invoices/1');
      expect(request.request.body).toEqual(withItem);

      request.flush({ ...withItem, id: 1, total: 301 });
    });

    it('replaces in the listing the invoice the API returned', () => {
      load();

      service.update(1, newInvoice).subscribe();
      http.expectOne('/api/billing/invoices/1').flush({ ...newInvoice, id: 1 });

      expect(service.invoices()).toHaveLength(2);
      expect(service.invoices()[0]).toMatchObject({ id: 1, number: 'NF-0100' });
    });

    it('leaves the other invoices alone', () => {
      load();

      service.update(1, newInvoice).subscribe();
      http.expectOne('/api/billing/invoices/1').flush({ ...newInvoice, id: 1 });

      expect(service.invoices()[1]).toEqual(invoices[1]);
    });

    it('does not mutate the previous list array', () => {
      load();
      const previousList = service.invoices();

      service.update(1, newInvoice).subscribe();
      http.expectOne('/api/billing/invoices/1').flush({ ...newInvoice, id: 1 });

      expect(service.invoices()).not.toBe(previousList);
      expect(previousList).toEqual(invoices);
    });

    it('changes nothing when the API rejects the invoice', () => {
      load();

      let failed = false;
      service.update(1, newInvoice).subscribe({ error: () => (failed = true) });
      http
        .expectOne('/api/billing/invoices/1')
        .flush(
          { errors: { number: 'Campo obrigatório.' } },
          { status: 400, statusText: 'Bad Request' }
        );

      expect(failed).toBe(true);
      expect(service.invoices()).toEqual(invoices);
    });
  });
});
