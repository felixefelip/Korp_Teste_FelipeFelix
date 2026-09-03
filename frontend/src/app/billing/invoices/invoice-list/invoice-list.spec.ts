import { HttpErrorResponse } from '@angular/common/http';
import { registerLocaleData } from '@angular/common';
import localePt from '@angular/common/locales/pt';
import { LOCALE_ID, signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { Observable, of, tap, throwError } from 'rxjs';

import { Invoice, STUCK_AFTER, UNSTABLE_AFTER } from '../invoice.model';
import { FlashService } from '../../../shared/flash/flash.service';
import { InvoiceService } from '../invoice.service';
import { POLL_INTERVAL, SLOW_POLL_INTERVAL, InvoiceList } from './invoice-list';

registerLocaleData(localePt, 'pt-BR');

const INVOICES: Invoice[] = [
  { id: 1, series: 1, number: 1, formattedNumber: '001/000001', type: 'OUT', status: 'OPEN', items: [], productsTotal: 4299.9, total: 4299.9, icmsBase: 4299.9, icmsValue: 773.98, ipiBase: 0, ipiValue: 0 },
  { id: 2, series: 1, number: 2, formattedNumber: '001/000002', type: 'OUT', status: 'CLOSED', items: [], productsTotal: 899, total: 899, icmsBase: 899, icmsValue: 161.82, ipiBase: 0, ipiValue: 0 },
  {
    id: 3,
    series: 2,
    number: 9999,
    formattedNumber: '002/009999',
    type: 'IN',
    status: 'OPEN',
    items: [],
    productsTotal: 0,
    total: 0,
    icmsBase: 0,
    icmsValue: 0,
    ipiBase: 0,
    ipiValue: 0
  }
];

describe('InvoiceList', () => {
  let fixture: ComponentFixture<InvoiceList>;
  let flash: { error: ReturnType<typeof vi.fn>; success: ReturnType<typeof vi.fn> };

  const text = (el: Element | null | undefined) =>
    (el?.textContent ?? '').replace(/[  ]/g, ' ').trim();

  const element = () => fixture.nativeElement as HTMLElement;

  const rows = () =>
    Array.from(element().querySelectorAll('tbody tr')).filter(
      (row) => !row.querySelector('.table__empty')
    );

  const cells = (row: Element) => Array.from(row.querySelectorAll('td')).map(text);

  const numbers = () => rows().map((row) => cells(row)[0]);

  const typeInFilter = async (term: string) => {
    const field = element().querySelector<HTMLInputElement>('input[type="search"]')!;
    field.value = term;
    field.dispatchEvent(new Event('input'));
    await fixture.whenStable();
  };

  const mount = async (response: () => Observable<Invoice[]>) => {
    TestBed.resetTestingModule();

    const invoices = signal<Invoice[]>([]);

    flash = { error: vi.fn(), success: vi.fn() };

    const service = {
      invoices,
      list: vi.fn(() => response().pipe(tap((list) => invoices.set(list))))
    };

    await TestBed.configureTestingModule({
      imports: [InvoiceList],
      providers: [
        provideRouter([]),
        { provide: InvoiceService, useValue: service },
        { provide: FlashService, useValue: flash },
        { provide: LOCALE_ID, useValue: 'pt-BR' }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(InvoiceList);
    await fixture.whenStable();

    return service;
  };

  beforeEach(async () => {
    await mount(() => of(INVOICES));
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('listing', () => {
    it('shows every invoice coming from the API', () => {
      expect(numbers()).toEqual(['001/000001', '001/000002', '002/009999']);
    });

    it('shows the invoice columns in the expected order', () => {
      expect(cells(rows()[0])).toEqual([
        '001/000001',
        'Saída',
        'Aberta',
        'R$ 4.299,90',
        'Editar'
      ]);
    });

    it('formats the total of each invoice in reais', () => {
      const totals = rows().map((row) => cells(row)[3]);

      expect(totals).toEqual(['R$ 4.299,90', 'R$ 899,00', 'R$ 0,00']);
    });

    it('translates the direction of each invoice', () => {
      const types = rows().map((row) => cells(row)[1]);
      expect(types).toEqual(['Saída', 'Saída', 'Entrada']);
    });

    it('translates the status of each invoice', () => {
      const statuses = rows().map((row) => cells(row)[2]);
      expect(statuses).toEqual(['Aberta', 'Fechada', 'Aberta']);
    });

    it('marks the status with the badge of its own state', () => {
      expect(rows()[0].querySelector('.badge--open')).not.toBeNull();
      expect(rows()[1].querySelector('.badge--closed')).not.toBeNull();
    });

    it('reports the invoice count in the subtitle', () => {
      expect(text(element().querySelector('.page__subtitle'))).toBe(
        '3 nota(s) fiscal(is) encontrada(s)'
      );
    });

    it('offers an edit action on each open invoice, pointing to its form', () => {
      const links = rows().map((row) =>
        row.querySelector('.menu-button__action')?.getAttribute('href')
      );

      expect(links).toEqual([
        '/billing/invoices/1/edit',
        null,
        '/billing/invoices/3/edit'
      ]);
    });

    it('never offers to edit a closed invoice', () => {
      expect(text(rows()[1].querySelector('.menu-button__action'))).toBe('Ver DANFE');
      expect(text(rows()[1])).not.toContain('Editar');
    });

    it('offers the create action pointing to the form', () => {
      const action = element().querySelector('a.btn--primary');

      expect(text(action)).toBe('+ Cadastrar nota fiscal');
      expect(action?.getAttribute('href')).toBe('/billing/invoices/new');
    });
  });

  describe('loading', () => {
    it('fetches the listing when the screen opens', () => {
      expect(TestBed.inject(InvoiceService).list).toHaveBeenCalled();
    });

    it('warns when the API does not answer', async () => {
      TestBed.resetTestingModule();
      await mount(() => throwError(() => new Error('network down')));

      expect(text(element().querySelector('.table__error'))).toContain(
        'Não foi possível carregar as notas fiscais.'
      );
      expect(rows()).toHaveLength(0);
    });

    it('retries when the user asks for it', async () => {
      TestBed.resetTestingModule();

      let shouldFail = true;
      await mount(() =>
        shouldFail ? throwError(() => new Error('network down')) : of(INVOICES)
      );

      shouldFail = false;
      element().querySelector<HTMLButtonElement>('.table__error button')!.click();
      await fixture.whenStable();

      expect(numbers()).toHaveLength(3);
    });
  });

  describe('filter', () => {
    it('filters by number', async () => {
      await typeInFilter('001/000002');
      expect(numbers()).toEqual(['001/000002']);
    });

    it('ignores case differences', async () => {
      await typeInFilter('001/000001');
      expect(numbers()).toEqual(['001/000001']);
    });

    it('ignores spaces around the term', async () => {
      await typeInFilter('   002/009999   ');
      expect(numbers()).toEqual(['002/009999']);
    });

    it('returns several invoices when the term is common', async () => {
      await typeInFilter('001/');
      expect(numbers()).toEqual(['001/000001', '001/000002']);
    });

    it('updates the subtitle count while filtering', async () => {
      await typeInFilter('001/');
      expect(text(element().querySelector('.page__subtitle'))).toBe(
        '2 nota(s) fiscal(is) encontrada(s)'
      );
    });

    it('shows the empty state when nothing matches', async () => {
      await typeInFilter('inexistente');

      expect(rows()).toHaveLength(0);
      expect(text(element().querySelector('.table__empty'))).toBe(
        'Nenhuma nota fiscal encontrada.'
      );
    });

    it('lists everything again when the filter is cleared', async () => {
      await typeInFilter('001/000001');
      await typeInFilter('');

      expect(numbers()).toHaveLength(3);
    });
  });

  describe('while an invoice is being processed', () => {
    const processing = (): Invoice[] => [{ ...INVOICES[0], status: 'CLOSING' }];
    const settled = (): Invoice[] => [{ ...INVOICES[0], status: 'CLOSED' }];

    it('translates the status as processing', async () => {
      await mount(() => of(processing()));

      expect(cells(rows()[0])[2]).toBe('Processando');
    });

    it('keeps asking the API while the invoice has not settled', async () => {
      vi.useFakeTimers({ shouldAdvanceTime: true });

      try {
        const service = await mount(() => of(processing()));
        const initial = service.list.mock.calls.length;

        await vi.advanceTimersByTimeAsync(POLL_INTERVAL * 3);

        expect(service.list.mock.calls.length).toBeGreaterThan(initial);
      } finally {
        vi.useRealTimers();
      }
    });

    it('stops asking once nothing is being processed', async () => {
      vi.useFakeTimers({ shouldAdvanceTime: true });

      try {
        let answered = false;

        const service = await mount(() => {
          const list = answered ? settled() : processing();
          answered = true;

          return of(list);
        });

        await vi.advanceTimersByTimeAsync(POLL_INTERVAL * 2);
        const afterSettling = service.list.mock.calls.length;

        await vi.advanceTimersByTimeAsync(POLL_INTERVAL * 3);

        expect(service.list.mock.calls.length).toBe(afterSettling);
      } finally {
        vi.useRealTimers();
      }
    });

    it('shows the reason when the stock refused the invoice', async () => {
      await mount(() =>
        of([{ ...INVOICES[0], status: 'OPEN' as const, failureReason: 'INSUFFICIENT_STOCK' }])
      );

      expect(text(element().querySelector('.table__failure'))).toBe('Estoque insuficiente');
    });
  });

  describe('while an invoice is being reopened', () => {
    const reopening = (): Invoice[] => [{ ...INVOICES[1], status: 'REOPENING' }];

    it('translates the status as processing too', async () => {
      await mount(() => of(reopening()));

      expect(cells(rows()[0])[2]).toBe('Processando');
    });

    it('keeps asking the API, the same as a closing invoice', async () => {
      vi.useFakeTimers({ shouldAdvanceTime: true });

      try {
        const service = await mount(() => of(reopening()));
        const initial = service.list.mock.calls.length;

        await vi.advanceTimersByTimeAsync(POLL_INTERVAL * 3);

        expect(service.list.mock.calls.length).toBeGreaterThan(initial);
      } finally {
        vi.useRealTimers();
      }
    });

    it('shows the reason when the stock could not be given back', async () => {
      await mount(() =>
        of([{ ...INVOICES[1], status: 'CLOSED' as const, failureReason: 'STOCK_ALREADY_USED' }])
      );

      expect(text(element().querySelector('.table__failure'))).toBe(
        'O estoque desta nota já foi utilizado'
      );
    });
  });

  describe('while an invoice takes longer than expected', () => {
    const processingFor = (elapsed: number): Invoice[] => [
      {
        ...INVOICES[0],
        status: 'CLOSING',
        processingSince: new Date(Date.now() - elapsed).toISOString()
      }
    ];

    const note = () => text(element().querySelector('.table__failure'));

    const action = () => text(element().querySelector('.menu-button__action'));

    it('says nothing extra in the first seconds', async () => {
      await mount(() => of(processingFor(1000)));

      expect(note()).toBe('');
      expect(element().querySelector('.menu-button')).toBeNull();
    });

    it('owns up to the instability once it drags on', async () => {
      await mount(() => of(processingFor(UNSTABLE_AFTER + 1000)));

      expect(note()).toContain('instabilidade');
      expect(note()).toContain('Aguarde alguns instantes');
    });

    it('offers no action while the retries are still running', async () => {
      await mount(() => of(processingFor(UNSTABLE_AFTER + 1000)));

      expect(element().querySelector('.menu-button')).toBeNull();
    });

    it('reports the failure once the retries are exhausted', async () => {
      await mount(() => of(processingFor(STUCK_AFTER + 1000)));

      expect(note()).toBe(
        'Não foi possível concluir o processamento desta nota fiscal. Tente novamente ou entre em contato com o suporte.'
      );
    });

    it('only then offers retrying', async () => {
      await mount(() => of(processingFor(STUCK_AFTER + 1000)));

      expect(action()).toBe('Tentar novamente');
    });

    it('says nothing about an invoice the API gave no clock for', async () => {
      await mount(() => of([{ ...INVOICES[0], status: 'CLOSING' as const }]));

      expect(note()).toBe('');
    });

    it('slows the polling down once nothing is about to settle', async () => {
      vi.useFakeTimers({ shouldAdvanceTime: true });

      try {
        const service = await mount(() => of(processingFor(STUCK_AFTER + 1000)));
        const initial = service.list.mock.calls.length;

        await vi.advanceTimersByTimeAsync(POLL_INTERVAL * 3);

        expect(service.list.mock.calls.length).toBe(initial);

        await vi.advanceTimersByTimeAsync(SLOW_POLL_INTERVAL);

        expect(service.list.mock.calls.length).toBeGreaterThan(initial);
      } finally {
        vi.useRealTimers();
      }
    });

    it('keeps asking often while an invoice may still settle', async () => {
      vi.useFakeTimers({ shouldAdvanceTime: true });

      try {
        const service = await mount(() => of(processingFor(1000)));
        const initial = service.list.mock.calls.length;

        await vi.advanceTimersByTimeAsync(POLL_INTERVAL * 2);

        expect(service.list.mock.calls.length).toBeGreaterThan(initial);
      } finally {
        vi.useRealTimers();
      }
    });
  });

  describe('reporting what was missing', () => {
    it('names the product that was short', async () => {
      await mount(() =>
        of([
          {
            ...INVOICES[0],
            status: 'OPEN' as const,
            failureReason: 'INSUFFICIENT_STOCK',
            shortages: [
              { inventoryId: 42, code: 'PROD-1', name: 'Parafuso', required: 50, available: 42 }
            ]
          }
        ])
      );

      expect(text(element().querySelector('.table__failure'))).toBe(
        'Estoque insuficiente: PROD-1'
      );
    });

    it('lists every product when more than one was short', async () => {
      await mount(() =>
        of([
          {
            ...INVOICES[0],
            status: 'OPEN' as const,
            failureReason: 'INSUFFICIENT_STOCK',
            shortages: [
              { inventoryId: 42, code: 'PROD-1', name: 'Parafuso', required: 50, available: 42 },
              { inventoryId: 43, code: 'PROD-2', name: 'Arruela', required: 5, available: 0 }
            ]
          }
        ])
      );

      expect(text(element().querySelector('.table__failure'))).toBe(
        'Estoque insuficiente: PROD-1, PROD-2'
      );
    });
  });
});
