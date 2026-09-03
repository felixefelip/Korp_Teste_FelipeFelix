import { provideHttpClient } from '@angular/common/http';
import {
  HttpTestingController,
  provideHttpClientTesting
} from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import {
  Invoice,
  InvoiceDocument,
  InvoiceDraft,
  InvoicePayload
} from './invoice.model';
import { InvoiceService } from './invoice.service';

describe('InvoiceService', () => {
  let service: InvoiceService;
  let http: HttpTestingController;

  const newInvoice: InvoicePayload = { series: 1, number: 100, type: 'OUT', items: [] };

  const withItem: InvoicePayload = {
    series: 1, number: 101,
    type: 'OUT',
    items: [
      {
        inventoryId: 3,
        code: 'PRD-0003',
        name: 'Cadeira',
        unit: 'UN',
        quantity: 2,
        unitPrice: 150.5,
        icmsRate: 18,
        ipiRate: 10
      }
    ]
  };

  const invoices: Invoice[] = [
    { id: 1, series: 1, number: 1, formattedNumber: '001/000001', type: 'OUT', status: 'OPEN', items: [], productsTotal: 0, total: 0, icmsBase: 0, icmsValue: 0, ipiBase: 0, ipiValue: 0 },
    { id: 2, series: 1, number: 2, formattedNumber: '001/000002', type: 'OUT', status: 'CLOSED', items: [], productsTotal: 0, total: 0, icmsBase: 0, icmsValue: 0, ipiBase: 0, ipiValue: 0 }
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
        .flush({ ...newInvoice, id: 7, series: 1, number: 100 });

      expect(service.invoices()).toHaveLength(3);
      expect(service.invoices().at(-1)).toMatchObject({ id: 7, series: 1, number: 100 });
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

      request.flush({ id: 7, series: 1, number: 7, formattedNumber: '001/000007', type: 'OUT', status: 'CLOSED', items: [], total: 0 });

      expect(received).toEqual({
        id: 7,
        series: 1,
        number: 7,
        formattedNumber: '001/000007',
        type: 'OUT',
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
        series: 1, number: 7,
        type: 'OUT',
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
      expect(service.invoices()[0]).toMatchObject({ id: 1, series: 1, number: 100 });
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

  describe('remove', () => {
    it('deletes that invoice', () => {
      service.remove(7).subscribe();

      const request = http.expectOne('/api/billing/invoices/7');

      expect(request.request.method).toBe('DELETE');
      request.flush(null);
    });

    it('takes it out of the listing', () => {
      load();

      service.remove(1).subscribe();
      http.expectOne('/api/billing/invoices/1').flush(null);

      expect(service.invoices().map((invoice) => invoice.id)).toEqual([2]);
    });

    it('leaves the listing alone when the API refuses', () => {
      load();

      service.remove(1).subscribe({ error: () => undefined });
      http.expectOne('/api/billing/invoices/1').flush('', { status: 409, statusText: 'Conflict' });

      expect(service.invoices()).toHaveLength(2);
    });
  });

  describe('close', () => {
    it('posts to the close action of that invoice', () => {
      service.close(7).subscribe();

      const request = http.expectOne('/api/billing/invoices/7/close');

      expect(request.request.method).toBe('POST');
      expect(request.request.body).toEqual({});
      request.flush({ ...invoices[0], id: 7, status: 'CLOSED' });
    });

    it('swaps the invoice for the closed one', () => {
      load();

      service.close(1).subscribe();
      http.expectOne('/api/billing/invoices/1/close').flush({ ...invoices[0], status: 'CLOSED' });

      expect(service.invoices()[0].status).toBe('CLOSED');
      expect(service.invoices()).toHaveLength(2);
    });
  });

  describe('reopen', () => {
    it('posts to the reopen action of that invoice', () => {
      service.reopen(7).subscribe();

      const request = http.expectOne('/api/billing/invoices/7/reopen');

      expect(request.request.method).toBe('POST');
      expect(request.request.body).toEqual({});
      request.flush({ ...invoices[1], id: 7, status: 'OPEN' });
    });

    it('swaps the invoice for the reopened one', () => {
      load();

      service.reopen(2).subscribe();
      http.expectOne('/api/billing/invoices/2/reopen').flush({
        ...invoices[1],
        status: 'OPEN'
      });

      expect(service.invoices()[1].status).toBe('OPEN');
      expect(service.invoices()).toHaveLength(2);
    });

    it('leaves the listing alone when the API refuses', () => {
      load();

      service.reopen(2).subscribe({ error: () => undefined });
      http
        .expectOne('/api/billing/invoices/2/reopen')
        .flush('', { status: 409, statusText: 'Conflict' });

      expect(service.invoices()[1].status).toBe('CLOSED');
    });
  });

  describe('retry', () => {
    it('posts to the retry action of that invoice', () => {
      service.retry(7).subscribe();

      const request = http.expectOne('/api/billing/invoices/7/retry');

      expect(request.request.method).toBe('POST');
      expect(request.request.body).toEqual({});
      request.flush({ ...invoices[1], id: 7, status: 'CLOSING' });
    });

    it('keeps the invoice in the listing, still being processed', () => {
      load();

      service.retry(2).subscribe();
      http.expectOne('/api/billing/invoices/2/retry').flush({
        ...invoices[1],
        status: 'CLOSING'
      });

      expect(service.invoices()[1].status).toBe('CLOSING');
      expect(service.invoices()).toHaveLength(2);
    });
  });

  describe('draft', () => {
    const draft: InvoiceDraft = {
      type: 'OUT',
      items: withItem.items,
      unresolved: []
    };

    it('posts the prompt to the draft endpoint', () => {
      service.draft('vender 2 cadeiras').subscribe();

      const request = http.expectOne('/api/billing/invoices/draft');

      expect(request.request.method).toBe('POST');
      expect(request.request.body).toEqual({ prompt: 'vender 2 cadeiras' });
      request.flush(draft);
    });

    it('returns the draft without touching the listing', () => {
      load();

      let received: InvoiceDraft | undefined;
      service.draft('vender 2 cadeiras').subscribe((value) => (received = value));
      http.expectOne('/api/billing/invoices/draft').flush(draft);

      expect(received).toEqual(draft);
      expect(service.invoices()).toEqual(invoices);
    });
  });

  describe('nextDocument', () => {
    const document: InvoiceDocument = { series: 1, number: 7 };

    it('asks for the suggestion of the last series in use', () => {
      let received: InvoiceDocument | undefined;
      service.nextDocument().subscribe((value) => (received = value));

      const request = http.expectOne('/api/billing/invoices/next-document');

      expect(request.request.method).toBe('GET');
      request.flush(document);

      expect(received).toEqual(document);
    });

    it('asks for the suggestion of a given series', () => {
      service.nextDocument(4).subscribe();

      http.expectOne('/api/billing/invoices/next-document?series=4').flush({
        series: 4,
        number: 1
      });
    });

    it('accepts a series with no number left', () => {
      let received: InvoiceDocument | undefined;
      service.nextDocument(1).subscribe((value) => (received = value));

      http
        .expectOne('/api/billing/invoices/next-document?series=1')
        .flush({ series: 1, number: null });

      expect(received?.number).toBeNull();
    });
  });
});
