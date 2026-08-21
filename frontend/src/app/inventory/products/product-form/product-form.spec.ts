import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { Product } from '../product.model';
import { ProductForm } from './product-form';
import { ProductPayload } from '../product.model';

const EXISTING: Product = {
  id: 7,
  code: 'PRD-0007',
  name: 'Cadeira de escritório',
  unit: 'CX',
  price: 750.5,
  stock: 8
};

describe('ProductForm', () => {
  let fixture: ComponentFixture<ProductForm>;
  let saved: ProductPayload[];

  const element = () => fixture.nativeElement as HTMLElement;

  const field = <T extends HTMLElement>(id: string) =>
    element().querySelector<T>(`#${id}`)!;

  const text = (el: Element | null | undefined) =>
    (el?.textContent ?? '').replace(/[\u00A0\u202F]/g, ' ').trim();

  const setInput = async (name: string, value: unknown) => {
    fixture.componentRef.setInput(name, value);
    await fixture.whenStable();
  };

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

  const banner = () => text(element().querySelector('.form__failure'));

  const submitButton = () =>
    element().querySelector<HTMLButtonElement>('button[type="submit"]')!;

  const fillValidForm = async () => {
    await fill('code', 'PRD-0006');
    await fill('name', 'Cadeira de escritório');
    await fill('unit', 'CX');
    await fill('price', '750.5');
    await fill('stock', '8');
  };

  const fillMinimum = async () => {
    await fill('code', 'PRD-0006');
    await fill('name', 'Cadeira de escritório');
    await fill('price', '750.5');
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ProductForm],
      providers: [provideRouter([])]
    }).compileComponents();

    fixture = TestBed.createComponent(ProductForm);
    saved = [];
    fixture.componentInstance.save.subscribe((product) => saved.push(product));
    await fixture.whenStable();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('initial state', () => {
    it('starts with an empty code', () => {
      expect(field<HTMLInputElement>('code').value).toBe('');
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
      expect(banner()).toBe('');
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

      expect(saved).toEqual([]);
      expect(errorOf('code')).toBe('Campo obrigatório.');
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
      expect(saved).toEqual([expect.objectContaining({ code: 'PRD-0001' })]);
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

      expect(saved).toEqual([expect.objectContaining({ price: 0, stock: 0 })]);
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

  describe('emitting what was filled', () => {
    it('hands over the filled data', async () => {
      await fillValidForm();
      await submit();

      expect(saved).toEqual([
        {
          code: 'PRD-0006',
          name: 'Cadeira de escritório',
          unit: 'CX',
          price: 750.5,
          stock: 8
        }
      ]);
    });

    it('trims surrounding spaces from the name', async () => {
      await fillValidForm();
      await fill('name', '   Mesa de reunião   ');
      await submit();

      expect(saved).toEqual([expect.objectContaining({ name: 'Mesa de reunião' })]);
    });

    it('hands over the default stock when the field is untouched', async () => {
      await fillMinimum();
      await submit();

      expect(saved).toEqual([expect.objectContaining({ stock: 0 })]);
    });

    it('emits only once per submit', async () => {
      await fillValidForm();
      await submit();

      expect(saved.length).toBe(1);
    });

    it('emits nothing while a save is already running', async () => {
      await fillValidForm();
      await setInput('saving', true);
      await submit();

      expect(saved).toEqual([]);
    });
  });

  describe('the product being edited', () => {
    it('fills every field with the value it was given', async () => {
      await setInput('value', EXISTING);

      expect(field<HTMLInputElement>('code').value).toBe('PRD-0007');
      expect(field<HTMLInputElement>('name').value).toBe('Cadeira de escritório');
      expect(field<HTMLSelectElement>('unit').value).toBe('CX');
      expect(field<HTMLInputElement>('price').value).toBe('750.5');
      expect(field<HTMLInputElement>('stock').value).toBe('8');
    });

    it('shows no error over a product that is still untouched', async () => {
      await setInput('value', EXISTING);

      expect(errors()).toEqual([]);
      expect(banner()).toBe('');
    });

    it('hands over the edited data', async () => {
      await setInput('value', EXISTING);
      await fill('name', 'Cadeira gamer');
      await fill('price', '899.9');
      await submit();

      expect(saved).toEqual([
        {
          code: 'PRD-0007',
          name: 'Cadeira gamer',
          unit: 'CX',
          price: 899.9,
          stock: 8
        }
      ]);
    });

    it('still blocks the submit when a field was emptied', async () => {
      await setInput('value', EXISTING);
      await fill('name', '');
      await submit();

      expect(saved).toEqual([]);
      expect(errorOf('name')).toBe('Campo obrigatório.');
    });
  });

  describe('labels and states given by the page', () => {
    it('names the submit button as asked', async () => {
      await setInput('submitLabel', 'Salvar alterações');

      expect(text(submitButton())).toBe('Salvar alterações');
    });

    it('names the stock field as asked', async () => {
      await setInput('stockLabel', 'Estoque inicial');

      expect(text(field('stock').parentElement?.querySelector('.field-label'))).toBe(
        'Estoque inicial'
      );
    });

    it('announces the save in progress', async () => {
      await setInput('saving', true);

      expect(text(submitButton())).toBe('Salvando…');
      expect(submitButton().disabled).toBe(true);
    });

    it('waits for the product before showing the fields', async () => {
      await setInput('loading', true);

      expect(text(element().querySelector('.form__status'))).toBe(
        'Carregando produto…'
      );
      expect(element().querySelector('#code')).toBeNull();
      expect(submitButton().disabled).toBe(true);
    });

    it('emits nothing while the product is still loading', async () => {
      await setInput('loading', true);
      await submit();

      expect(saved).toEqual([]);
    });
  });

  describe('failure coming from the page', () => {
    const rejectWith = (fieldErrors: unknown, message = 'Não foi possível salvar.') =>
      setInput('failure', { fieldErrors, message });

    it('shows on the field the error the server pointed at', async () => {
      await fillValidForm();
      await rejectWith({ code: 'Campo obrigatório.' });

      expect(errorOf('code')).toBe('Campo obrigatório.');
      expect(banner()).toBe('');
    });

    it('shows the server phrase exactly as it came', async () => {
      await fillValidForm();
      await rejectWith({ price: 'O valor não pode ser negativo.' });

      expect(errorOf('price')).toBe('O valor não pode ser negativo.');
    });

    it('shows a phrase the frontend has no copy for', async () => {
      await fillValidForm();
      await rejectWith({ unit: 'Esta unidade foi descontinuada.' });

      expect(errorOf('unit')).toBe('Esta unidade foi descontinuada.');
    });

    it('drops the server error as soon as the field is fixed', async () => {
      await fillValidForm();
      await rejectWith({ code: 'Campo obrigatório.' });
      expect(errorOf('code')).toBe('Campo obrigatório.');

      await fill('code', 'PRD-0042');
      expect(errorOf('code')).toBe('');
    });

    it('shows a general warning when no field was pointed at', async () => {
      await fillValidForm();
      await rejectWith(null, 'Não foi possível ler os dados enviados.');

      expect(banner()).toBe('Não foi possível ler os dados enviados.');
      expect(errors()).toEqual([]);
    });

    it('shows a general warning when the pointed field does not exist here', async () => {
      await fillValidForm();
      await rejectWith({ supplier: 'Fornecedor inválido.' }, 'Dados inválidos.');

      expect(banner()).toBe('Dados inválidos.');
      expect(errors()).toEqual([]);
    });

    it('drops the warning when the page starts another save', async () => {
      await fillValidForm();
      await rejectWith(null, 'Não foi possível salvar.');
      expect(banner()).toBe('Não foi possível salvar.');

      await setInput('failure', null);

      expect(banner()).toBe('');
    });
  });
});
