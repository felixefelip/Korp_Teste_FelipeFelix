import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';

import { Invoice } from '../invoice.model';
import { InvoiceService } from '../invoice.service';
import { InvoiceForm } from './invoice-form';

describe('InvoiceForm', () => {
  let fixture: ComponentFixture<InvoiceForm>;
  let service: {
    create: ReturnType<typeof vi.fn>;
  };
  let navigate: ReturnType<typeof vi.spyOn>;

  const element = () => fixture.nativeElement as HTMLElement;

  const field = <T extends HTMLElement>(id: string) =>
    element().querySelector<T>(`#${id}`)!;

  const text = (el: Element | null | undefined) =>
    (el?.textContent ?? '').replace(/[  ]/g, ' ').trim();

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

  const rejectWith = (body: unknown, status = 400) =>
    service.create.mockReturnValue(
      throwError(() => new HttpErrorResponse({ status, error: body }))
    );

  const fillValidForm = async () => {
    await fill('number', 'NF-0006');
    await fill('status', 'CLOSED');
  };

  beforeEach(async () => {
    service = {
      create: vi.fn((data: Omit<Invoice, 'id'>) => of({ ...data, id: 6 }))
    };

    await TestBed.configureTestingModule({
      imports: [InvoiceForm],
      providers: [provideRouter([]), { provide: InvoiceService, useValue: service }]
    }).compileComponents();

    navigate = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);

    fixture = TestBed.createComponent(InvoiceForm);
    await fixture.whenStable();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('initial state', () => {
    it('starts with an empty number', () => {
      expect(field<HTMLInputElement>('number').value).toBe('');
    });

    it('starts with the invoice open', () => {
      expect(field<HTMLSelectElement>('status').value).toBe('OPEN');
    });

    it('lists the available statuses in Portuguese', () => {
      const options = Array.from(field<HTMLSelectElement>('status').options).map(
        (option) => [option.value, text(option)]
      );

      expect(options).toEqual([
        ['OPEN', 'Aberta'],
        ['CLOSED', 'Fechada']
      ]);
    });

    it('shows no error before any interaction', () => {
      expect(errors()).toEqual([]);
      expect(failure()).toBe('');
    });

    it('offers cancel going back to the listing', () => {
      expect(element().querySelector('a.btn--ghost')?.getAttribute('href')).toBe(
        '/billing/invoices'
      );
    });
  });

  describe('validation', () => {
    it('blocks the submit and reveals the required field errors', async () => {
      await submit();

      expect(service.create).not.toHaveBeenCalled();
      expect(navigate).not.toHaveBeenCalled();
      expect(errorOf('number')).toBe('Campo obrigatório.');
      expect(errorOf('status')).toBe('');
    });

    it('demands the number once the field is emptied', async () => {
      await fill('number', 'NF-0006');
      await fill('number', '');

      expect(errorOf('number')).toBe('Campo obrigatório.');
    });

    it('rejects a number longer than the 30 characters the API accepts', async () => {
      await fill('number', 'N'.repeat(31));

      expect(errorOf('number')).toBe('Limite de 30 caracteres excedido.');
    });

    it('accepts a number with exactly 30 characters', async () => {
      await fill('number', 'N'.repeat(30));
      await submit();

      expect(errorOf('number')).toBe('');
      expect(service.create).toHaveBeenCalled();
    });

    it('accepts a number already used by another invoice', async () => {
      await fillValidForm();
      await fill('number', 'NF-0001');
      await submit();

      expect(errorOf('number')).toBe('');
      expect(service.create).toHaveBeenCalledWith(
        expect.objectContaining({ number: 'NF-0001' })
      );
    });

    it('marks the invalid field visually', async () => {
      await submit();
      expect(field('number').classList).toContain('field--error');
    });

    it('clears the error as soon as the field is fixed', async () => {
      await submit();
      expect(errorOf('number')).toBe('Campo obrigatório.');

      await fill('number', 'NF-0006');
      expect(errorOf('number')).toBe('');
    });
  });

  describe('creation', () => {
    it('sends the filled data to the service', async () => {
      await fillValidForm();
      await submit();

      expect(service.create).toHaveBeenCalledWith({
        number: 'NF-0006',
        status: 'CLOSED'
      });
    });

    it('trims surrounding spaces from the number', async () => {
      await fill('number', '   NF-0007   ');
      await submit();

      expect(service.create).toHaveBeenCalledWith(
        expect.objectContaining({ number: 'NF-0007' })
      );
    });

    it('saves as open when the status is untouched', async () => {
      await fill('number', 'NF-0006');
      await submit();

      expect(service.create).toHaveBeenCalledWith(
        expect.objectContaining({ status: 'OPEN' })
      );
    });

    it('goes back to the listing after saving', async () => {
      await fillValidForm();
      await submit();

      expect(navigate).toHaveBeenCalledWith(['/billing/invoices']);
    });

    it('creates only once per submit', async () => {
      await fillValidForm();
      await submit();

      expect(service.create).toHaveBeenCalledTimes(1);
    });
  });

  describe('API rejection', () => {
    it('shows on the field the error the server pointed at', async () => {
      rejectWith({ errors: { number: 'Campo obrigatório.' } });

      await fillValidForm();
      await submit();

      expect(errorOf('number')).toBe('Campo obrigatório.');
      expect(navigate).not.toHaveBeenCalled();
    });

    it('shows the server phrase exactly as it came', async () => {
      rejectWith({ errors: { number: 'Limite de 30 caracteres excedido.' } });

      await fillValidForm();
      await submit();

      expect(errorOf('number')).toBe('Limite de 30 caracteres excedido.');
    });

    it('shows a phrase the frontend has no copy for', async () => {
      rejectWith({ errors: { status: 'Esta nota já foi transmitida.' } });

      await fillValidForm();
      await submit();

      expect(errorOf('status')).toBe('Esta nota já foi transmitida.');
    });

    it('drops the server error as soon as the field is fixed', async () => {
      rejectWith({ errors: { number: 'Campo obrigatório.' } });

      await fillValidForm();
      await submit();
      expect(errorOf('number')).toBe('Campo obrigatório.');

      await fill('number', 'NF-0042');
      expect(errorOf('number')).toBe('');
    });

    it('shows a general warning when the body points at no field', async () => {
      rejectWith({ message: 'Não foi possível ler os dados enviados.' });

      await fillValidForm();
      await submit();

      expect(failure()).toBe('Não foi possível ler os dados enviados.');
      expect(navigate).not.toHaveBeenCalled();
    });

    it('accepts field errors on a 422', async () => {
      rejectWith({ errors: { number: 'Campo obrigatório.' } }, 422);

      await fillValidForm();
      await submit();

      expect(errorOf('number')).toBe('Campo obrigatório.');
    });

    it('shows a general warning when the API is down', async () => {
      rejectWith(null, 500);

      await fillValidForm();
      await submit();

      expect(failure()).toBe('Não foi possível salvar a nota fiscal. Tente novamente.');
    });

    it('never shows the body message of a 500 to the user', async () => {
      rejectWith({ message: 'erro ao criar a nota fiscal' }, 500);

      await fillValidForm();
      await submit();

      expect(failure()).toBe('Não foi possível salvar a nota fiscal. Tente novamente.');
      expect(navigate).not.toHaveBeenCalled();
    });

    it('does not mark fields from the body of a 500', async () => {
      rejectWith({ errors: { number: 'Campo obrigatório.' } }, 500);

      await fillValidForm();
      await submit();

      expect(errorOf('number')).toBe('');
      expect(failure()).toBe('Não foi possível salvar a nota fiscal. Tente novamente.');
    });

    it('shows a general warning when there is no response at all', async () => {
      rejectWith(new ProgressEvent('error'), 0);

      await fillValidForm();
      await submit();

      expect(failure()).toBe('Não foi possível salvar a nota fiscal. Tente novamente.');
    });

    it('allows fixing and trying again', async () => {
      rejectWith({ errors: { number: 'Campo obrigatório.' } });

      await fillValidForm();
      await submit();

      service.create.mockReturnValue(of({ id: 6 } as Invoice));
      await fill('number', 'NF-0042');
      await submit();

      expect(service.create).toHaveBeenCalledTimes(2);
      expect(navigate).toHaveBeenCalledWith(['/billing/invoices']);
    });
  });
});
