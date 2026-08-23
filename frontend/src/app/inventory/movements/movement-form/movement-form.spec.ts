import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { Movement, MovementPayload } from '../movement.model';
import { MovementForm } from './movement-form';

const EXISTING: Movement = {
  id: 4,
  productId: 7,
  type: 'out',
  origin: 'adjustment',
  quantity: 3,
  confirmed: true,
  billingInvoiceItemId: null,
  billingInvoiceId: null
};

const BACK_LINK = ['/inventory/products', 7, 'movements'];

describe('MovementForm', () => {
  let fixture: ComponentFixture<MovementForm>;
  let saved: MovementPayload[];

  const element = () => fixture.nativeElement as HTMLElement;

  const field = <T extends HTMLElement>(id: string) =>
    element().querySelector<T>(`#${id}`)!;

  const text = (el: Element | null | undefined) =>
    (el?.textContent ?? '').replace(/[  ]/g, ' ').trim();

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

  const check = async (id: string, value: boolean) => {
    const input = field<HTMLInputElement>(id);
    input.checked = value;
    input.dispatchEvent(new Event('change'));
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

  const banner = () => text(element().querySelector('.form__failure'));

  const submitButton = () =>
    element().querySelector<HTMLButtonElement>('button[type="submit"]')!;

  beforeEach(async () => {
    saved = [];

    await TestBed.configureTestingModule({
      imports: [MovementForm],
      providers: [provideRouter([])]
    }).compileComponents();

    fixture = TestBed.createComponent(MovementForm);
    fixture.componentRef.setInput('backLink', BACK_LINK);
    fixture.componentInstance.save.subscribe((movement) => saved.push(movement));
    await fixture.whenStable();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('the empty form', () => {
    it('starts as an unconfirmed entry', () => {
      expect(field<HTMLSelectElement>('type').value).toBe('in');
      expect(field<HTMLInputElement>('confirmed').checked).toBe(false);
      expect(field<HTMLInputElement>('quantity').value).toBe('');
    });

    it('offers both directions, in portuguese', () => {
      const options = Array.from(field<HTMLSelectElement>('type').options).map(
        (option) => option.text.trim()
      );

      expect(options).toEqual(['Entrada', 'Saída']);
    });

    it('points cancel back where the container asked', () => {
      expect(
        element().querySelector<HTMLAnchorElement>('.btn--ghost')!.getAttribute('href')
      ).toBe('/inventory/products/7/movements');
    });
  });

  describe('filling it in', () => {
    it('emits what was typed', async () => {
      await fill('type', 'out');
      await fill('quantity', '5');
      await check('confirmed', true);
      await submit();

      expect(saved).toEqual([{ type: 'out', quantity: 5, confirmed: true }]);
    });

    it('emits an unconfirmed movement as such', async () => {
      await fill('quantity', '5');
      await submit();

      expect(saved).toEqual([{ type: 'in', quantity: 5, confirmed: false }]);
    });
  });

  describe('validation', () => {
    it('refuses a movement without a quantity', async () => {
      await submit();

      expect(saved).toEqual([]);
      expect(errorOf('quantity')).toBe('Campo obrigatório.');
    });

    it('refuses a quantity of zero, which moves nothing', async () => {
      await fill('quantity', '0');
      await submit();

      expect(saved).toEqual([]);
      expect(errorOf('quantity')).toBe('O valor mínimo é 1.');
    });

    it('refuses a negative quantity, since the type carries the sign', async () => {
      await fill('quantity', '-3');
      await submit();

      expect(saved).toEqual([]);
      expect(errorOf('quantity')).toBe('O valor mínimo é 1.');
    });

    it('refuses a fractional quantity', async () => {
      await fill('quantity', '1.5');
      await submit();

      expect(saved).toEqual([]);
      expect(errorOf('quantity')).toBe('Informe um número inteiro.');
    });

    it('shows no error before the first submit', () => {
      expect(errorOf('quantity')).toBe('');
    });
  });

  describe('an existing movement', () => {
    it('fills every field with it', async () => {
      await setInput('value', EXISTING);

      expect(field<HTMLSelectElement>('type').value).toBe('out');
      expect(field<HTMLInputElement>('quantity').value).toBe('3');
      expect(field<HTMLInputElement>('confirmed').checked).toBe(true);
    });

    it('emits the edited values', async () => {
      await setInput('value', EXISTING);
      await fill('quantity', '9');
      await submit();

      expect(saved).toEqual([{ type: 'out', quantity: 9, confirmed: true }]);
    });
  });

  describe('the states the container drives', () => {
    it('waits for the movement before showing the fields', async () => {
      await setInput('loading', true);

      expect(element().querySelector('#quantity')).toBeNull();
      expect(text(element().querySelector('.form__status'))).toBe(
        'Carregando movimentação…'
      );
      expect(submitButton().disabled).toBe(true);
    });

    it('says it is saving', async () => {
      await setInput('saving', true);

      expect(text(submitButton())).toBe('Salvando…');
      expect(submitButton().disabled).toBe(true);
    });

    it('lets the container name the submit button', async () => {
      await setInput('submitLabel', 'Salvar alterações');

      expect(text(submitButton())).toBe('Salvar alterações');
    });
  });

  describe('an error coming from the server', () => {
    it('lands on the field it names', async () => {
      await fill('quantity', '5');
      await submit();
      await setInput('failure', {
        fieldErrors: { quantity: 'Quantidade acima do saldo.' },
        message: 'Não foi possível salvar a movimentação. Tente novamente.'
      });

      expect(errorOf('quantity')).toBe('Quantidade acima do saldo.');
      expect(banner()).toBe('');
    });

    it('falls back to the banner when it names no field', async () => {
      await setInput('failure', {
        fieldErrors: null,
        message: 'Movimentações geradas por notas fiscais não podem ser alteradas.'
      });

      expect(banner()).toBe(
        'Movimentações geradas por notas fiscais não podem ser alteradas.'
      );
    });
  });
});
