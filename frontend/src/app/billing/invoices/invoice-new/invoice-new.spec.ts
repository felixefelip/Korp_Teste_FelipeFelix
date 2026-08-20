import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';
import { Subject, of, throwError } from 'rxjs';

import { Invoice } from '../invoice.model';
import { InvoiceService } from '../invoice.service';
import { InvoiceNew } from './invoice-new';

describe('InvoiceNew', () => {
  let fixture: ComponentFixture<InvoiceNew>;
  let service: {
    create: ReturnType<typeof vi.fn>;
    get: ReturnType<typeof vi.fn>;
    update: ReturnType<typeof vi.fn>;
  };
  let navigate: ReturnType<typeof vi.spyOn>;

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

  const errorOf = (id: string) =>
    text(field(id).parentElement?.querySelector('.field-error'));

  const failure = () => text(element().querySelector('.form__failure'));

  const submitButton = () =>
    element().querySelector<HTMLButtonElement>('button[type="submit"]')!;

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
      create: vi.fn((data: Omit<Invoice, 'id'>) => of({ ...data, id: 6 })),
      get: vi.fn(),
      update: vi.fn()
    };

    await TestBed.configureTestingModule({
      imports: [InvoiceNew],
      providers: [provideRouter([]), { provide: InvoiceService, useValue: service }]
    }).compileComponents();

    navigate = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);

    fixture = TestBed.createComponent(InvoiceNew);
    await fixture.whenStable();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  it('announces that it is creating, not editing', () => {
    expect(text(element().querySelector('.page__title'))).toBe(
      'Cadastrar nota fiscal'
    );
    expect(text(submitButton())).toBe('Salvar nota fiscal');
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

    it('never creates from an invalid form', async () => {
      await submit();

      expect(service.create).not.toHaveBeenCalled();
      expect(navigate).not.toHaveBeenCalled();
    });

    it('blocks a second submit while the first is still running', async () => {
      service.create.mockReturnValue(new Subject<Invoice>());

      await fillValidForm();
      await submit();
      await submit();

      expect(service.create).toHaveBeenCalledTimes(1);
      expect(text(submitButton())).toBe('Salvando…');
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

    it('drops the warning of the previous attempt', async () => {
      rejectWith(null, 500);

      await fillValidForm();
      await submit();
      expect(failure()).toBe('Não foi possível salvar a nota fiscal. Tente novamente.');

      service.create.mockReturnValue(new Subject<Invoice>());
      await submit();

      expect(failure()).toBe('');
    });
  });
});
