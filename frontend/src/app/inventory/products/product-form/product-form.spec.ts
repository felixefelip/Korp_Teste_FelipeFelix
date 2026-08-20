import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';

import { Product } from '../product.model';
import { ProductService } from '../product.service';
import { ProductForm } from './product-form';

describe('ProductForm', () => {
  let fixture: ComponentFixture<ProductForm>;
  let service: {
    nextCode: ReturnType<typeof vi.fn>;
    list: ReturnType<typeof vi.fn>;
    create: ReturnType<typeof vi.fn>;
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
    await fill('name', 'Cadeira de escritório');
    await fill('unit', 'CX');
    await fill('price', '750.5');
    await fill('stock', '8');
  };

  const fillMinimum = async () => {
    await fill('name', 'Cadeira de escritório');
    await fill('price', '750.5');
  };

  beforeEach(async () => {
    service = {
      nextCode: vi.fn().mockReturnValue('PRD-0006'),
      list: vi.fn(() => of([] as Product[])),
      create: vi.fn((data: Omit<Product, 'id'>) => of({ ...data, id: 6 }))
    };

    await TestBed.configureTestingModule({
      imports: [ProductForm],
      providers: [provideRouter([]), { provide: ProductService, useValue: service }]
    }).compileComponents();

    navigate = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);

    fixture = TestBed.createComponent(ProductForm);
    await fixture.whenStable();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('initial state', () => {
    it('suggests the next available code', () => {
      expect(service.nextCode).toHaveBeenCalled();
      expect(field<HTMLInputElement>('code').value).toBe('PRD-0006');
    });

    it('loads the listing so it can suggest the code', () => {
      expect(service.list).toHaveBeenCalled();
    });

    it('starts with the initial stock zeroed', () => {
      expect(field<HTMLInputElement>('stock').value).toBe('0');
    });

    it('starts with the first unit selected', () => {
      expect(field<HTMLSelectElement>('unit').value).toBe('UN');
    });

    it('lists the available units of measure', () => {
      const options = Array.from(field<HTMLSelectElement>('unit').options).map(
        (option) => option.value
      );
      expect(options).toEqual(['UN', 'CX', 'PC', 'KG', 'L', 'M']);
    });

    it('shows no error before any interaction', () => {
      expect(errors()).toEqual([]);
      expect(failure()).toBe('');
    });

    it('offers cancel going back to the listing', () => {
      expect(element().querySelector('a.btn--ghost')?.getAttribute('href')).toBe(
        '/inventory/products'
      );
    });
  });

  describe('validation', () => {
    it('blocks the submit and reveals the required field errors', async () => {
      await submit();

      expect(service.create).not.toHaveBeenCalled();
      expect(navigate).not.toHaveBeenCalled();
      expect(errorOf('name')).toBe('Campo obrigatório.');
      expect(errorOf('price')).toBe('Campo obrigatório.');
      expect(errorOf('stock')).toBe('');
    });

    it('demands the stock once the field is emptied', async () => {
      await fill('stock', '');
      expect(errorOf('stock')).toBe('Campo obrigatório.');
    });

    it('requires a name with at least 3 characters', async () => {
      await fill('name', 'ab');
      expect(errorOf('name')).toBe('Informe pelo menos 3 caracteres.');
    });

    it('accepts a code already used by another product', async () => {
      await fillValidForm();
      await fill('code', 'PRD-0001');
      await submit();

      expect(errorOf('code')).toBe('');
      expect(service.create).toHaveBeenCalledWith(
        expect.objectContaining({ code: 'PRD-0001' })
      );
    });

    it('rejects a negative price', async () => {
      await fill('price', '-1');
      expect(errorOf('price')).toBe('O valor não pode ser negativo.');
    });

    it('rejects a negative stock', async () => {
      await fill('stock', '-5');
      expect(errorOf('stock')).toBe('O valor não pode ser negativo.');
    });

    it('rejects a fractional stock', async () => {
      await fill('stock', '2.5');
      expect(errorOf('stock')).toBe('Informe um número inteiro.');
    });

    it('accepts zeroed price and stock', async () => {
      await fillValidForm();
      await fill('price', '0');
      await fill('stock', '0');
      await submit();

      expect(service.create).toHaveBeenCalledWith(
        expect.objectContaining({ price: 0, stock: 0 })
      );
    });

    it('marks the invalid field visually', async () => {
      await submit();
      expect(field('name').classList).toContain('field--error');
    });

    it('clears the error as soon as the field is fixed', async () => {
      await submit();
      expect(errorOf('name')).toBe('Campo obrigatório.');

      await fill('name', 'Cadeira de escritório');
      expect(errorOf('name')).toBe('');
    });
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

    it('trims surrounding spaces from the name', async () => {
      await fillValidForm();
      await fill('name', '   Mesa de reunião   ');
      await submit();

      expect(service.create).toHaveBeenCalledWith(
        expect.objectContaining({ name: 'Mesa de reunião' })
      );
    });

    it('saves with the default stock when the field is untouched', async () => {
      await fillMinimum();
      await submit();

      expect(service.create).toHaveBeenCalledWith(
        expect.objectContaining({ stock: 0 })
      );
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
  });

  describe('API rejection', () => {
    it('shows on the field the error the server pointed at', async () => {
      rejectWith({ errors: { code: 'Campo obrigatório.' } });

      await fillValidForm();
      await submit();

      expect(errorOf('code')).toBe('Campo obrigatório.');
      expect(navigate).not.toHaveBeenCalled();
    });

    it('shows the server phrase exactly as it came', async () => {
      rejectWith({ errors: { price: 'O valor não pode ser negativo.' } });

      await fillValidForm();
      await submit();

      expect(errorOf('price')).toBe('O valor não pode ser negativo.');
    });

    it('shows a phrase the frontend has no copy for', async () => {
      rejectWith({ errors: { unit: 'Esta unidade foi descontinuada.' } });

      await fillValidForm();
      await submit();

      expect(errorOf('unit')).toBe('Esta unidade foi descontinuada.');
    });

    it('drops the server error as soon as the field is fixed', async () => {
      rejectWith({ errors: { code: 'Campo obrigatório.' } });

      await fillValidForm();
      await submit();
      expect(errorOf('code')).toBe('Campo obrigatório.');

      await fill('code', 'PRD-0042');
      expect(errorOf('code')).toBe('');
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
  });
});
