import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { Observable, of, tap, throwError } from 'rxjs';

import { Invoice } from '../invoice.model';
import { InvoiceService } from '../invoice.service';
import { InvoiceList } from './invoice-list';

const INVOICES: Invoice[] = [
  { id: 1, number: 'NF-0001', status: 'OPEN' },
  { id: 2, number: 'NF-0002', status: 'CLOSED' },
  { id: 3, number: 'ABC-9999', status: 'OPEN' }
];

describe('InvoiceList', () => {
  let fixture: ComponentFixture<InvoiceList>;

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
    const invoices = signal<Invoice[]>([]);

    const service = {
      invoices,
      list: vi.fn(() => response().pipe(tap((list) => invoices.set(list))))
    };

    await TestBed.configureTestingModule({
      imports: [InvoiceList],
      providers: [provideRouter([]), { provide: InvoiceService, useValue: service }]
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
      expect(numbers()).toEqual(['NF-0001', 'NF-0002', 'ABC-9999']);
    });

    it('shows the invoice columns in the expected order', () => {
      expect(cells(rows()[0])).toEqual(['NF-0001', 'Aberta', 'Editar']);
    });

    it('translates the status of each invoice', () => {
      const statuses = rows().map((row) => cells(row)[1]);
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

    it('offers an edit action per invoice pointing to its form', () => {
      const links = rows().map((row) =>
        row.querySelector('.table__action')?.getAttribute('href')
      );

      expect(links).toEqual([
        '/billing/invoices/1/edit',
        '/billing/invoices/2/edit',
        '/billing/invoices/3/edit'
      ]);
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
      await typeInFilter('NF-0002');
      expect(numbers()).toEqual(['NF-0002']);
    });

    it('ignores case differences', async () => {
      await typeInFilter('nf-0001');
      expect(numbers()).toEqual(['NF-0001']);
    });

    it('ignores spaces around the term', async () => {
      await typeInFilter('   ABC   ');
      expect(numbers()).toEqual(['ABC-9999']);
    });

    it('returns several invoices when the term is common', async () => {
      await typeInFilter('NF-');
      expect(numbers()).toEqual(['NF-0001', 'NF-0002']);
    });

    it('updates the subtitle count while filtering', async () => {
      await typeInFilter('NF-');
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
      await typeInFilter('NF-0001');
      await typeInFilter('');

      expect(numbers()).toHaveLength(3);
    });
  });
});
