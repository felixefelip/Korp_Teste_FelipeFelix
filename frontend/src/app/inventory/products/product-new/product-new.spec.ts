import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';
import { Subject, of, throwError } from 'rxjs';

import { Product } from '../product.model';
import { ProductService } from '../product.service';
import { ProductNew } from './product-new';

describe('ProductNew', () => {
  let fixture: ComponentFixture<ProductNew>;
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
    await fill('code', 'PRD-0006');
    await fill('name', 'Cadeira de escritório');
    await fill('unit', 'CX');
    await fill('price', '750.5');
    await fill('stock', '8');
  };

  beforeEach(async () => {
    service = {
      create: vi.fn((data: Omit<Product, 'id'>) => of({ ...data, id: 6 })),
      get: vi.fn(),
      update: vi.fn()
    };

    await TestBed.configureTestingModule({
      imports: [ProductNew],
      providers: [provideRouter([]), { provide: ProductService, useValue: service }]
    }).compileComponents();

    navigate = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);

    fixture = TestBed.createComponent(ProductNew);
    await fixture.whenStable();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  it('announces that it is creating, not editing', () => {
    expect(text(element().querySelector('.page__title'))).toBe('Cadastrar produto');
    expect(text(submitButton())).toBe('Salvar produto');
  });

  it('shows the stock as the initial one', () => {
    expect(text(field('stock').parentElement?.querySelector('.field-label'))).toBe(
      'Estoque inicial'
    );
  });

  describe('creation', () => {
    it('sends the filled data to the service', async () => {
      await fillValidForm();
      await submit();

      expect(service.create).toHaveBeenCalledWith({
        code: 'PRD-0006',
        name: 'Cadeira de escritório',
        unit: 'CX',
        price: 750.5,
        stock: 8
      });
    });

    it('goes back to the listing after saving', async () => {
      await fillValidForm();
      await submit();

      expect(navigate).toHaveBeenCalledWith(['/inventory/products']);
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
      service.create.mockReturnValue(new Subject<Product>());

      await fillValidForm();
      await submit();
      await submit();

      expect(service.create).toHaveBeenCalledTimes(1);
      expect(text(submitButton())).toBe('Salvando…');
    });
  });

  describe('API rejection', () => {
    it('shows on the field the error the server pointed at', async () => {
      rejectWith({ errors: { code: 'Campo obrigatório.' } });

      await fillValidForm();
      await submit();

      expect(errorOf('code')).toBe('Campo obrigatório.');
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
      rejectWith({ errors: { stock: 'O valor não pode ser negativo.' } }, 422);

      await fillValidForm();
      await submit();

      expect(errorOf('stock')).toBe('O valor não pode ser negativo.');
    });

    it('shows a general warning when the API is down', async () => {
      rejectWith(null, 500);

      await fillValidForm();
      await submit();

      expect(failure()).toBe('Não foi possível salvar o produto. Tente novamente.');
    });

    it('never shows the body message of a 500 to the user', async () => {
      rejectWith({ message: 'erro ao criar o produto' }, 500);

      await fillValidForm();
      await submit();

      expect(failure()).toBe('Não foi possível salvar o produto. Tente novamente.');
      expect(navigate).not.toHaveBeenCalled();
    });

    it('does not mark fields from the body of a 500', async () => {
      rejectWith({ errors: { code: 'Campo obrigatório.' } }, 500);

      await fillValidForm();
      await submit();

      expect(errorOf('code')).toBe('');
      expect(failure()).toBe('Não foi possível salvar o produto. Tente novamente.');
    });

    it('shows a general warning when there is no response at all', async () => {
      rejectWith(new ProgressEvent('error'), 0);

      await fillValidForm();
      await submit();

      expect(failure()).toBe('Não foi possível salvar o produto. Tente novamente.');
    });

    it('allows fixing and trying again', async () => {
      rejectWith({ errors: { code: 'Campo obrigatório.' } });

      await fillValidForm();
      await submit();

      service.create.mockReturnValue(of({ id: 6 } as Product));
      await fill('code', 'PRD-0042');
      await submit();

      expect(service.create).toHaveBeenCalledTimes(2);
      expect(navigate).toHaveBeenCalledWith(['/inventory/products']);
    });

    it('drops the warning of the previous attempt', async () => {
      rejectWith(null, 500);

      await fillValidForm();
      await submit();
      expect(failure()).toBe('Não foi possível salvar o produto. Tente novamente.');

      service.create.mockReturnValue(new Subject<Product>());
      await submit();

      expect(failure()).toBe('');
    });
  });
});
