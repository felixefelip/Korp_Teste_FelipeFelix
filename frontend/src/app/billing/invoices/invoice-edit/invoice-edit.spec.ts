import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router, convertToParamMap, provideRouter } from '@angular/router';
import { Observable, Subject, of, throwError } from 'rxjs';

import { FlashService } from '../../../shared/flash/flash.service';
import { Invoice } from '../invoice.model';
import { InvoiceService } from '../invoice.service';
import { InvoiceEdit } from './invoice-edit';

const EXISTING: Invoice = { id: 7, number: 'NF-0007', status: 'OPEN' };

describe('InvoiceEdit', () => {
  let fixture: ComponentFixture<InvoiceEdit>;
  let service: {
    create: ReturnType<typeof vi.fn>;
    get: ReturnType<typeof vi.fn>;
    update: ReturnType<typeof vi.fn>;
  };
  let navigate: ReturnType<typeof vi.spyOn>;
  let flash: { error: ReturnType<typeof vi.fn>; success: ReturnType<typeof vi.fn> };

  const element = () => fixture.nativeElement as HTMLElement;

  const field = <T extends HTMLElement>(id: string) =>
    element().querySelector<T>(`#${id}`)!;

  const text = (el: Element | null | undefined) =>
    (el?.textContent ?? '').replace(/[\u00A0\u202F]/g, ' ').trim();

  const fill = async (id: string, value: string) => {
    const input = field<HTMLInputElement | HTMLSelectElement>(id);
    input.value = value;
    input.dispatchEvent(new Event('input'));
    input.dispatchEvent(new Event('change'));
    input.dispatchEvent(new Event('blur'));
    await fixture.whenStable();
  };

  const submit = async () => {
    element()
      .querySelector('form')!
      .dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    await fixture.whenStable();
  };

  const errors = () =>
    Array.from(element().querySelectorAll('.field-error')).map(text);

  const errorOf = (id: string) =>
    text(field(id).parentElement?.querySelector('.field-error'));

  const failure = () => text(element().querySelector('.form__failure'));

  const submitButton = () =>
    element().querySelector<HTMLButtonElement>('button[type="submit"]')!;

  const mount = async (
    load: () => Observable<Invoice> = () => of(EXISTING),
    id = '7'
  ) => {
    TestBed.resetTestingModule();

    service = {
      create: vi.fn(),
      get: vi.fn(load),
      update: vi.fn((invoiceId: number, data: Omit<Invoice, 'id'>) =>
        of({ ...data, id: invoiceId })
      )
    };

    flash = { error: vi.fn(), success: vi.fn() };

    await TestBed.configureTestingModule({
      imports: [InvoiceEdit],
      providers: [
        provideRouter([]),
        { provide: InvoiceService, useValue: service },
        { provide: FlashService, useValue: flash },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: convertToParamMap({ id }) } }
        }
      ]
    }).compileComponents();

    navigate = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);

    fixture = TestBed.createComponent(InvoiceEdit);
    await fixture.whenStable();
  };

  beforeEach(async () => {
    await mount();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('loading the invoice', () => {
    it('asks the API for the invoice in the route', () => {
      expect(service.get).toHaveBeenCalledWith(7);
    });

    it('fills every field with what came back', () => {
      expect(field<HTMLInputElement>('number').value).toBe('NF-0007');
      expect(field<HTMLSelectElement>('status').value).toBe('OPEN');
    });

    it('announces that it is editing, not creating', () => {
      expect(text(element().querySelector('.page__title'))).toBe(
        'Editar nota fiscal'
      );
      expect(text(submitButton())).toBe('Salvar alterações');
    });

    it('shows no error over an invoice that is still untouched', () => {
      expect(errors()).toEqual([]);
      expect(failure()).toBe('');
      expect(flash.error).not.toHaveBeenCalled();
    });

    it('waits for the invoice before showing the fields', async () => {
      await mount(() => new Subject<Invoice>());

      expect(text(element().querySelector('.form__status'))).toBe(
        'Carregando nota fiscal…'
      );
      expect(element().querySelector('#number')).toBeNull();
      expect(submitButton().disabled).toBe(true);
    });
  });

  describe('saving the changes', () => {
    it('sends the edited data to the update of that id', async () => {
      await fill('number', 'NF-0042');
      await fill('status', 'CLOSED');
      await submit();

      expect(service.update).toHaveBeenCalledWith(7, {
        number: 'NF-0042',
        status: 'CLOSED'
      });
    });

    it('never creates a second invoice', async () => {
      await submit();

      expect(service.create).not.toHaveBeenCalled();
      expect(service.update).toHaveBeenCalledTimes(1);
    });

    it('trims surrounding spaces from the number', async () => {
      await fill('number', '   NF-0099   ');
      await submit();

      expect(service.update).toHaveBeenCalledWith(
        7,
        expect.objectContaining({ number: 'NF-0099' })
      );
    });

    it('goes back to the listing after saving', async () => {
      await submit();

      expect(navigate).toHaveBeenCalledWith(['/billing/invoices']);
    });

    it('still blocks the submit when a field was emptied', async () => {
      await fill('number', '');
      await submit();

      expect(service.update).not.toHaveBeenCalled();
      expect(errorOf('number')).toBe('Campo obrigatório.');
    });

    it('shows on the field the error the server pointed at', async () => {
      service.update.mockReturnValue(
        throwError(
          () =>
            new HttpErrorResponse({
              status: 400,
              error: { errors: { number: 'Campo obrigatório.' } }
            })
        )
      );

      await submit();

      expect(errorOf('number')).toBe('Campo obrigatório.');
      expect(navigate).not.toHaveBeenCalled();
    });

    it('warns without leaking the message of a 500', async () => {
      service.update.mockReturnValue(
        throwError(
          () =>
            new HttpErrorResponse({
              status: 500,
              error: { message: 'erro ao atualizar a nota fiscal' }
            })
        )
      );

      await submit();

      expect(failure()).toBe('Não foi possível salvar a nota fiscal. Tente novamente.');
      expect(navigate).not.toHaveBeenCalled();
    });
  });

  describe('invoice that cannot be loaded', () => {
    const failLoadWith = (status: number) =>
      mount(() => throwError(() => new HttpErrorResponse({ status })));

    it('goes back to the listing when the invoice does not exist', async () => {
      await failLoadWith(404);

      expect(navigate).toHaveBeenCalledWith(['/billing/invoices']);
    });

    it('goes back to the listing when the API is down', async () => {
      await failLoadWith(500);

      expect(navigate).toHaveBeenCalledWith(['/billing/invoices']);
    });

    it('says on the way out that the invoice does not exist', async () => {
      await failLoadWith(404);

      expect(flash.error).toHaveBeenCalledWith('Nota fiscal não encontrada.');
    });

    it('blames the API when the failure was not a 404', async () => {
      await failLoadWith(500);

      expect(flash.error).toHaveBeenCalledWith(
        'Não foi possível carregar a nota fiscal. Tente novamente.'
      );
    });

    it('warns once per failed load', async () => {
      await failLoadWith(404);

      expect(flash.error).toHaveBeenCalledTimes(1);
    });

    it('never shows the empty form while it leaves the screen', async () => {
      await failLoadWith(404);

      expect(element().querySelector('#number')).toBeNull();
      expect(text(element().querySelector('.form__status'))).toBe(
        'Carregando nota fiscal…'
      );
    });

    it('saves nothing on the way out', async () => {
      await failLoadWith(404);

      await submit();

      expect(service.update).not.toHaveBeenCalled();
      expect(service.create).not.toHaveBeenCalled();
    });
  });
});
