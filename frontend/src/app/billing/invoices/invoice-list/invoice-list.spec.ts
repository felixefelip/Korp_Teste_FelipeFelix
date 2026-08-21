import { HttpErrorResponse } from '@angular/common/http';
import { registerLocaleData } from '@angular/common';
import localePt from '@angular/common/locales/pt';
import { LOCALE_ID, signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { Observable, of, tap, throwError } from 'rxjs';

import { Invoice } from '../invoice.model';
import { FlashService } from '../../../shared/flash/flash.service';
import { InvoiceService } from '../invoice.service';
import { InvoiceList } from './invoice-list';

registerLocaleData(localePt, 'pt-BR');

const INVOICES: Invoice[] = [
  { id: 1, number: 'NF-0001', type: 'OUT', status: 'OPEN', items: [], total: 4299.9 },
  { id: 2, number: 'NF-0002', type: 'OUT', status: 'CLOSED', items: [], total: 899 },
  { id: 3, number: 'ABC-9999', type: 'IN', status: 'OPEN', items: [], total: 0 }
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

  const mount = async (
    response: () => Observable<Invoice[]>,
    remove: () => Observable<void> = () => of(undefined),
    close: () => Observable<Invoice> = () => of({ ...INVOICES[0], status: 'CLOSED' }),
    reopen: () => Observable<Invoice> = () => of({ ...INVOICES[1], status: 'OPEN' })
  ) => {
    TestBed.resetTestingModule();

    const invoices = signal<Invoice[]>([]);

    flash = { error: vi.fn(), success: vi.fn() };

    const service = {
      invoices,
      list: vi.fn(() => response().pipe(tap((list) => invoices.set(list)))),
      remove: vi.fn((id: number) =>
        remove().pipe(tap(() => invoices.update((list) => list.filter((i) => i.id !== id))))
      ),
      close: vi.fn(() =>
        close().pipe(
          tap((closed) =>
            invoices.update((list) => list.map((i) => (i.id === closed.id ? closed : i)))
          )
        )
      ),
      reopen: vi.fn(() =>
        reopen().pipe(
          tap((opened) =>
            invoices.update((list) => list.map((i) => (i.id === opened.id ? opened : i)))
          )
        )
      )
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
      expect(numbers()).toEqual(['NF-0001', 'NF-0002', 'ABC-9999']);
    });

    it('shows the invoice columns in the expected order', () => {
      expect(cells(rows()[0])).toEqual([
        'NF-0001',
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
      expect(text(rows()[1].querySelector('.menu-button__action'))).toBe('Reabrir');
      expect(rows()[1].querySelector('.menu-button__toggle')).toBeNull();
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

  const menuLabels = () =>
    Array.from(element().querySelectorAll('.menu-button__item')).map(text);

  describe('deleting an invoice', () => {
    const openMenuOf = async (index: number) => {
      rows()[index].querySelector<HTMLButtonElement>('.menu-button__toggle')?.click();
      await fixture.whenStable();
    };

    const openDelete = async (index: number) => {
      await openMenuOf(index);

      Array.from(element().querySelectorAll<HTMLButtonElement>('button.menu-button__item'))
        .find((item) => text(item) === 'Excluir')!
        .click();
      await fixture.whenStable();
    };

    const dialog = () => element().querySelector('.confirm');

    const dialogButton = (label: string) =>
      Array.from(element().querySelectorAll<HTMLButtonElement>('.confirm__actions .btn')).find(
        (button) => text(button).startsWith(label)
      )!;

    it('offers it on an open invoice', async () => {
      await mount(() => of(INVOICES));
      await openMenuOf(0);

      expect(menuLabels()).toEqual(['Imprimir', 'Excluir']);
    });

    it('never offers it on a closed invoice', async () => {
      await mount(() => of(INVOICES));

      expect(text(rows()[1].querySelector('.menu-button__action'))).toBe('Reabrir');
      expect(rows()[1].querySelector('.menu-button__toggle')).toBeNull();
    });

    it('asks before deleting anything', async () => {
      const service = await mount(() => of(INVOICES));
      await openDelete(0);

      expect(text(dialog())).toContain('NF-0001');
      expect(service.remove).not.toHaveBeenCalled();
    });

    it('deletes the invoice that was chosen', async () => {
      const service = await mount(() => of(INVOICES));
      await openDelete(0);

      dialogButton('Excluir').click();
      await fixture.whenStable();

      expect(service.remove).toHaveBeenCalledWith(1);
      expect(rows()).toHaveLength(2);
      expect(flash.success).toHaveBeenCalledWith('Nota fiscal NF-0001 excluída.');
    });

    it('deletes nothing when it is cancelled', async () => {
      const service = await mount(() => of(INVOICES));
      await openDelete(0);

      dialogButton('Cancelar').click();
      await fixture.whenStable();

      expect(service.remove).not.toHaveBeenCalled();
      expect(rows()).toHaveLength(3);
    });

    it('shows what the API said when it refuses', async () => {
      await mount(
        () => of(INVOICES),
        () =>
          throwError(
            () =>
              new HttpErrorResponse({
                status: 409,
                error: { message: 'Notas fiscais fechadas não podem ser excluídas.' }
              })
          )
      );
      await openDelete(0);

      dialogButton('Excluir').click();
      await fixture.whenStable();

      expect(rows()).toHaveLength(3);
      expect(flash.error).toHaveBeenCalledWith(
        'Notas fiscais fechadas não podem ser excluídas.'
      );
    });

    it('falls back to its own message when the API names none', async () => {
      await mount(
        () => of(INVOICES),
        () => throwError(() => new HttpErrorResponse({ status: 500 }))
      );
      await openDelete(0);

      dialogButton('Excluir').click();
      await fixture.whenStable();

      expect(flash.error).toHaveBeenCalledWith(
        'Não foi possível excluir a nota fiscal. Tente novamente.'
      );
    });
  });

  describe('printing an invoice', () => {
    const openMenuOf = async (index: number) => {
      rows()[index].querySelector<HTMLButtonElement>('.menu-button__toggle')?.click();
      await fixture.whenStable();
    };

    const openPrint = async (index: number) => {
      await openMenuOf(index);

      Array.from(element().querySelectorAll<HTMLButtonElement>('button.menu-button__item'))
        .find((item) => text(item) === 'Imprimir')!
        .click();
      await fixture.whenStable();
    };

    const dialog = () => element().querySelector('.confirm');

    const dialogButton = (label: string) =>
      Array.from(element().querySelectorAll<HTMLButtonElement>('.confirm__actions .btn')).find(
        (button) => text(button).startsWith(label)
      )!;

    it('warns that it settles the invoice', async () => {
      const service = await mount(() => of(INVOICES));
      await openPrint(0);

      expect(text(dialog())).toContain('NF-0001');
      expect(text(dialog())).toContain('fechada');
      expect(service.close).not.toHaveBeenCalled();
    });

    it('closes the invoice that was chosen', async () => {
      const service = await mount(() => of(INVOICES));
      await openPrint(0);

      dialogButton('Imprimir').click();
      await fixture.whenStable();

      expect(service.close).toHaveBeenCalledWith(1);
      expect(flash.success).toHaveBeenCalledWith('Nota fiscal NF-0001 fechada.');
    });

    it('shows the new status on the row', async () => {
      await mount(() => of(INVOICES));
      await openPrint(0);

      dialogButton('Imprimir').click();
      await fixture.whenStable();

      expect(cells(rows()[0])[2]).toBe('Fechada');
    });

    it('leaves the row offering nothing but to reopen afterwards', async () => {
      await mount(() => of(INVOICES));
      await openPrint(0);

      dialogButton('Imprimir').click();
      await fixture.whenStable();

      expect(text(rows()[0].querySelector('.menu-button__action'))).toBe('Reabrir');
      expect(rows()[0].querySelector('.menu-button__toggle')).toBeNull();
    });

    it('closes nothing when it is cancelled', async () => {
      const service = await mount(() => of(INVOICES));
      await openPrint(0);

      dialogButton('Cancelar').click();
      await fixture.whenStable();

      expect(service.close).not.toHaveBeenCalled();
      expect(dialog()).toBeNull();
    });

    it('shows what the API said when it refuses', async () => {
      await mount(
        () => of(INVOICES),
        () => of(undefined),
        () =>
          throwError(
            () =>
              new HttpErrorResponse({
                status: 409,
                error: { message: 'Esta nota fiscal já está fechada.' }
              })
          )
      );
      await openPrint(0);

      dialogButton('Imprimir').click();
      await fixture.whenStable();

      expect(flash.error).toHaveBeenCalledWith('Esta nota fiscal já está fechada.');
      expect(cells(rows()[0])[2]).toBe('Aberta');
    });

    it('falls back to its own message when the API names none', async () => {
      await mount(
        () => of(INVOICES),
        () => of(undefined),
        () => throwError(() => new HttpErrorResponse({ status: 500 }))
      );
      await openPrint(0);

      dialogButton('Imprimir').click();
      await fixture.whenStable();

      expect(flash.error).toHaveBeenCalledWith(
        'Não foi possível imprimir a nota fiscal. Tente novamente.'
      );
    });
  });

  describe('reopening an invoice', () => {
    const openMenuOf = async (index: number) => {
      rows()[index].querySelector<HTMLButtonElement>('.menu-button__toggle')?.click();
      await fixture.whenStable();
    };

    const chooseReopen = async (index: number) => {
      rows()[index].querySelector<HTMLButtonElement>('.menu-button__action')!.click();
      await fixture.whenStable();
    };

    const mountWith = (reopen: () => Observable<Invoice>) =>
      mount(
        () => of(INVOICES),
        () => of(undefined),
        () => of({ ...INVOICES[0], status: 'CLOSED' as const }),
        reopen
      );

    it('reopens without asking, since it settles nothing', async () => {
      const service = await mount(() => of(INVOICES));
      await chooseReopen(1);

      expect(element().querySelector('.confirm')).toBeNull();
      expect(service.reopen).toHaveBeenCalledWith(2);
      expect(flash.success).toHaveBeenCalledWith('Nota fiscal NF-0002 reaberta.');
    });

    it('shows the new status on the row', async () => {
      await mount(() => of(INVOICES));
      await chooseReopen(1);

      expect(cells(rows()[1])[2]).toBe('Aberta');
    });

    it('gives the row its full menu back', async () => {
      await mount(() => of(INVOICES));
      await chooseReopen(1);

      expect(text(rows()[1].querySelector('.menu-button__action'))).toBe('Editar');

      await openMenuOf(1);

      expect(menuLabels()).toEqual(['Imprimir', 'Excluir']);
    });

    it('shows what the API said when it refuses', async () => {
      await mountWith(() =>
        throwError(
          () =>
            new HttpErrorResponse({
              status: 409,
              error: { message: 'Esta nota fiscal já está aberta.' }
            })
        )
      );
      await chooseReopen(1);

      expect(flash.error).toHaveBeenCalledWith('Esta nota fiscal já está aberta.');
      expect(cells(rows()[1])[2]).toBe('Fechada');
    });

    it('falls back to its own message when the API names none', async () => {
      await mountWith(() => throwError(() => new HttpErrorResponse({ status: 500 })));
      await chooseReopen(1);

      expect(flash.error).toHaveBeenCalledWith(
        'Não foi possível reabrir a nota fiscal. Tente novamente.'
      );
    });
  });
});
