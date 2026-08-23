import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { Observable, Subject, of, throwError } from 'rxjs';

import { FlashService } from '../../../shared/flash/flash.service';
import { Invoice } from '../invoice.model';
import { InvoiceService } from '../invoice.service';
import { InvoiceActions } from './invoice-actions';

const OPEN: Invoice = {
  id: 7,
  series: 1, number: 7, formattedNumber: '001/000007',
  type: 'OUT',
  status: 'OPEN',
  items: [],
  total: 0
};

const CLOSED: Invoice = { ...OPEN, status: 'CLOSED' };
const PROCESSING: Invoice = { ...OPEN, status: 'CLOSING' };
const REOPENING: Invoice = { ...OPEN, status: 'REOPENING' };

describe('InvoiceActions', () => {
  let fixture: ComponentFixture<InvoiceActions>;
  let service: {
    close: ReturnType<typeof vi.fn>;
    reopen: ReturnType<typeof vi.fn>;
    retry: ReturnType<typeof vi.fn>;
    remove: ReturnType<typeof vi.fn>;
    danfeUrl: ReturnType<typeof vi.fn>;
  };
  let openWindow: ReturnType<typeof vi.spyOn>;
  let flash: { error: ReturnType<typeof vi.fn>; success: ReturnType<typeof vi.fn> };

  const element = () => fixture.nativeElement as HTMLElement;

  const text = (el: Element | null | undefined) =>
    (el?.textContent ?? '').replace(/\s+/g, ' ').trim();

  const primary = () => element().querySelector('.menu-button__action')!;

  const toggle = () => element().querySelector<HTMLButtonElement>('.menu-button__toggle');

  const openMenu = async () => {
    toggle()!.click();
    await fixture.whenStable();
  };

  const menuLabels = () =>
    Array.from(element().querySelectorAll('.menu-button__item')).map(text);

  const choose = async (label: string) => {
    await openMenu();

    Array.from(element().querySelectorAll<HTMLButtonElement>('button.menu-button__item'))
      .find((item) => text(item) === label)!
      .click();
    await fixture.whenStable();
  };

  const dialog = () => element().querySelector('.confirm');

  const dialogButton = (label: string) =>
    Array.from(element().querySelectorAll<HTMLButtonElement>('.confirm__actions .btn')).find(
      (button) => text(button).startsWith(label)
    )!;

  const click = async (el: Element) => {
    (el as HTMLElement).click();
    await fixture.whenStable();
  };

  const mount = async (
    invoice: Invoice = OPEN,
    responses: {
      close?: () => Observable<Invoice>;
      reopen?: () => Observable<Invoice>;
      retry?: () => Observable<Invoice>;
      remove?: () => Observable<void>;
    } = {}
  ) => {
    TestBed.resetTestingModule();

    service = {
      close: vi.fn(responses.close ?? (() => of(CLOSED))),
      reopen: vi.fn(responses.reopen ?? (() => of(OPEN))),
      retry: vi.fn(responses.retry ?? (() => of(PROCESSING))),
      remove: vi.fn(responses.remove ?? (() => of(undefined))),
      danfeUrl: vi.fn((id: number) => `/api/billing/invoices/${id}/danfe`)
    };

    openWindow = vi.spyOn(window, 'open').mockReturnValue(null);

    flash = { error: vi.fn(), success: vi.fn() };

    await TestBed.configureTestingModule({
      imports: [InvoiceActions],
      providers: [
        provideRouter([]),
        { provide: InvoiceService, useValue: service },
        { provide: FlashService, useValue: flash }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(InvoiceActions);
    fixture.componentRef.setInput('invoice', invoice);
    await fixture.whenStable();
  };

  beforeEach(async () => {
    await mount();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('an open invoice', () => {
    it('offers editing as the button itself', () => {
      expect(text(primary())).toBe('Editar');
      expect(primary().getAttribute('href')).toBe('/billing/invoices/7/edit');
    });

    it('offers printing and deleting behind the chevron', async () => {
      await openMenu();

      expect(menuLabels()).toEqual(['Imprimir', 'Excluir']);
    });
  });

  describe('a closed invoice', () => {
    beforeEach(async () => {
      await mount(CLOSED);
    });

    it('offers the danfe as the button itself', () => {
      expect(text(primary())).toBe('Ver DANFE');
      expect(primary().getAttribute('href')).toBeNull();
    });

    it('keeps reopening behind the chevron, since it gives the stock back', async () => {
      await openMenu();

      expect(menuLabels()).toEqual(['Reabrir']);
    });

    it('never offers editing or deleting', () => {
      expect(element().textContent).not.toContain('Editar');
      expect(element().textContent).not.toContain('Excluir');
    });
  });

  describe('printing', () => {
    it('asks before closing anything', async () => {
      await choose('Imprimir');

      expect(text(dialog())).toContain('001/000007');
      expect(service.close).not.toHaveBeenCalled();
    });

    it('closes the invoice and says so', async () => {
      await choose('Imprimir');
      await click(dialogButton('Imprimir'));

      expect(service.close).toHaveBeenCalledWith(7);
      expect(flash.success).toHaveBeenCalledWith('Nota fiscal 001/000007 fechada.');
      expect(dialog()).toBeNull();
    });

    it('closes nothing when cancelled', async () => {
      await choose('Imprimir');
      await click(dialogButton('Cancelar'));

      expect(service.close).not.toHaveBeenCalled();
      expect(dialog()).toBeNull();
    });

    it('says it is working while the API does not answer', async () => {
      await mount(OPEN, { close: () => new Subject<Invoice>() });
      await choose('Imprimir');
      await click(dialogButton('Imprimir'));

      expect(text(dialogButton('Imprimindo'))).toBe('Imprimindo…');
      expect(dialog()).not.toBeNull();
    });

    it('shows what the API said when it refuses', async () => {
      await mount(OPEN, {
        close: () =>
          throwError(
            () =>
              new HttpErrorResponse({
                status: 409,
                error: { message: 'Esta nota fiscal já está fechada.' }
              })
          )
      });
      await choose('Imprimir');
      await click(dialogButton('Imprimir'));

      expect(flash.error).toHaveBeenCalledWith('Esta nota fiscal já está fechada.');
      expect(dialog()).toBeNull();
    });

    it('falls back to its own message when the API names none', async () => {
      await mount(OPEN, {
        close: () => throwError(() => new HttpErrorResponse({ status: 500 }))
      });
      await choose('Imprimir');
      await click(dialogButton('Imprimir'));

      expect(flash.error).toHaveBeenCalledWith(
        'Não foi possível imprimir a nota fiscal. Tente novamente.'
      );
    });
  });

  describe('deleting', () => {
    it('asks before deleting anything', async () => {
      await choose('Excluir');

      expect(text(dialog())).toContain('001/000007');
      expect(service.remove).not.toHaveBeenCalled();
    });

    it('deletes the invoice and says so', async () => {
      await choose('Excluir');
      await click(dialogButton('Excluir'));

      expect(service.remove).toHaveBeenCalledWith(7);
      expect(flash.success).toHaveBeenCalledWith('Nota fiscal 001/000007 excluída.');
    });

    it('deletes nothing when cancelled', async () => {
      await choose('Excluir');
      await click(dialogButton('Cancelar'));

      expect(service.remove).not.toHaveBeenCalled();
      expect(dialog()).toBeNull();
    });

    it('shows what the API said when it refuses', async () => {
      await mount(OPEN, {
        remove: () =>
          throwError(
            () =>
              new HttpErrorResponse({
                status: 409,
                error: { message: 'Notas fiscais fechadas não podem ser excluídas.' }
              })
          )
      });
      await choose('Excluir');
      await click(dialogButton('Excluir'));

      expect(flash.error).toHaveBeenCalledWith(
        'Notas fiscais fechadas não podem ser excluídas.'
      );
    });

    it('falls back to its own message when the API names none', async () => {
      await mount(OPEN, {
        remove: () => throwError(() => new HttpErrorResponse({ status: 500 }))
      });
      await choose('Excluir');
      await click(dialogButton('Excluir'));

      expect(flash.error).toHaveBeenCalledWith(
        'Não foi possível excluir a nota fiscal. Tente novamente.'
      );
    });
  });

  describe('viewing the danfe', () => {
    beforeEach(async () => {
      await mount(CLOSED);
    });

    it('opens the danfe of the invoice in another tab', async () => {
      await click(primary());

      expect(service.danfeUrl).toHaveBeenCalledWith(7);
      expect(openWindow).toHaveBeenCalledWith(
        '/api/billing/invoices/7/danfe',
        '_blank',
        'noopener'
      );
    });

    it('asks nothing and changes nothing, since opening the danfe only reads', async () => {
      await click(primary());

      expect(dialog()).toBeNull();
      expect(service.close).not.toHaveBeenCalled();
      expect(service.reopen).not.toHaveBeenCalled();
      expect(flash.error).not.toHaveBeenCalled();
    });
  });

  describe('reopening', () => {
    it('reopens without asking, since it settles nothing', async () => {
      await mount(CLOSED);
      await choose('Reabrir');

      expect(dialog()).toBeNull();
      expect(service.reopen).toHaveBeenCalledWith(7);
      expect(flash.success).toHaveBeenCalledWith('Nota fiscal 001/000007 reaberta.');
    });

    it('shows what the API said when it refuses', async () => {
      await mount(CLOSED, {
        reopen: () =>
          throwError(
            () =>
              new HttpErrorResponse({
                status: 409,
                error: { message: 'Esta nota fiscal já está aberta.' }
              })
          )
      });
      await choose('Reabrir');

      expect(flash.error).toHaveBeenCalledWith('Esta nota fiscal já está aberta.');
    });

    it('falls back to its own message when the API names none', async () => {
      await mount(CLOSED, {
        reopen: () => throwError(() => new HttpErrorResponse({ status: 500 }))
      });
      await choose('Reabrir');

      expect(flash.error).toHaveBeenCalledWith(
        'Não foi possível reabrir a nota fiscal. Tente novamente.'
      );
    });
  });

  describe('following the invoice it is given', () => {
    it('swaps its actions when the invoice closes under it', async () => {
      expect(text(primary())).toBe('Editar');

      fixture.componentRef.setInput('invoice', CLOSED);
      await fixture.whenStable();

      expect(text(primary())).toBe('Ver DANFE');
    });
  });

  describe('while the invoice is being processed', () => {
    beforeEach(async () => {
      await mount(PROCESSING);
    });

    it('offers retrying as the button itself, for the result that never came', () => {
      expect(text(primary())).toBe('Tentar novamente');
    });

    it('offers nothing else behind a chevron', () => {
      expect(toggle()).toBeNull();
    });

    it('does not offer printing again', () => {
      expect(element().textContent).not.toContain('Imprimir');
    });

    it('does not offer reopening, which only makes sense once it is closed', () => {
      expect(element().textContent).not.toContain('Reabrir');
    });

    it('does not offer editing or deleting', () => {
      expect(element().textContent).not.toContain('Editar');
      expect(element().textContent).not.toContain('Excluir');
    });
  });

  describe('while the invoice is being reopened', () => {
    beforeEach(async () => {
      await mount(REOPENING);
    });

    it('offers retrying as well', () => {
      expect(text(primary())).toBe('Tentar novamente');
    });

    it('does not offer reopening again', () => {
      expect(element().textContent).not.toContain('Reabrir');
    });
  });

  describe('retrying', () => {
    it('resends the request without asking, since it settles nothing', async () => {
      await mount(PROCESSING);
      await click(primary());

      expect(dialog()).toBeNull();
      expect(service.retry).toHaveBeenCalledWith(7);
      expect(flash.success).toHaveBeenCalledWith('Nota fiscal 001/000007 reenviada.');
    });

    it('shows what the API said when it refuses', async () => {
      await mount(PROCESSING, {
        retry: () =>
          throwError(
            () =>
              new HttpErrorResponse({
                status: 409,
                error: { message: 'Esta nota fiscal não está em processamento.' }
              })
          )
      });
      await click(primary());

      expect(flash.error).toHaveBeenCalledWith('Esta nota fiscal não está em processamento.');
    });

    it('falls back to its own message when the API names none', async () => {
      await mount(PROCESSING, {
        retry: () => throwError(() => new HttpErrorResponse({ status: 500 }))
      });
      await click(primary());

      expect(flash.error).toHaveBeenCalledWith(
        'Não foi possível reenviar a nota fiscal. Tente novamente.'
      );
    });
  });
});
