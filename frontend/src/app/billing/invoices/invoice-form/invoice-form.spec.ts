import { registerLocaleData } from '@angular/common';
import localePt from '@angular/common/locales/pt';
import { LOCALE_ID } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { Product } from '../../../inventory/products/product.model';
import { Invoice, InvoiceItemPayload, InvoicePayload } from '../invoice.model';
import { InvoiceForm } from './invoice-form';

registerLocaleData(localePt, 'pt-BR');

const PRODUCTS: Product[] = [
  {
    id: 3,
    code: 'PRD-0003',
    name: 'Cadeira Gamer',
    unit: 'UN',
    price: 150.5,
    stock: 10
  },
  {
    id: 5,
    code: 'PRD-0005',
    name: 'Mesa de Escritório',
    unit: 'CX',
    price: 899,
    stock: 4
  }
];

const EXISTING_ITEM: InvoiceItemPayload = {
  inventoryId: 3,
  code: 'PRD-0003',
  name: 'Cadeira Gamer',
  unit: 'UN',
  quantity: 2,
  unitPrice: 150.5
};

const EXISTING: Invoice = {
  id: 7,
  series: 1, number: 7, formattedNumber: '001/000007',
  type: 'OUT',
  status: 'CLOSED',
  total: 301,
  items: [{ ...EXISTING_ITEM, id: 1, productId: 11, total: 301 }]
};

describe('InvoiceForm', () => {
  let fixture: ComponentFixture<InvoiceForm>;
  let saved: InvoicePayload[];

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

  const addButton = () =>
    element().querySelector<HTMLButtonElement>('.items__header button')!;

  const addItem = async () => {
    addButton().click();
    await fixture.whenStable();
  };

  const removeItem = async (row: number) => {
    element().querySelectorAll<HTMLButtonElement>('.items__remove')[row].click();
    await fixture.whenStable();
  };

  const chooseProduct = async (row: number, productIndex: number) => {
    const select = field<HTMLSelectElement>(`item-${row}-product`);
    select.selectedIndex = productIndex + 1;
    select.dispatchEvent(new Event('change'));
    select.dispatchEvent(new Event('blur'));
    await fixture.whenStable();
  };

  const rows = () => Array.from(element().querySelectorAll('.items__table tbody tr'));

  const rowTotals = () =>
    Array.from(element().querySelectorAll('.items__row-total')).map(text);

  const invoiceTotal = () => text(element().querySelector('.items__total'));

  const unitOf = (row: number) => text(rows()[row].querySelector('.items__unit'));

  const fillValidForm = async () => {
    await fill('series', '1');
    await fill('number', '6');
  };

  const addValidItem = async (row = 0, productIndex = 0, quantity = '2') => {
    await addItem();
    await chooseProduct(row, productIndex);
    await fill(`item-${row}-quantity`, quantity);
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [InvoiceForm],
      providers: [provideRouter([]), { provide: LOCALE_ID, useValue: 'pt-BR' }]
    }).compileComponents();

    fixture = TestBed.createComponent(InvoiceForm);
    saved = [];
    fixture.componentInstance.save.subscribe((invoice) => saved.push(invoice));
    fixture.componentRef.setInput('products', PRODUCTS);
    await fixture.whenStable();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('initial state', () => {
    it('starts with an empty number', () => {
      expect(field<HTMLInputElement>('number').value).toBe('');
    });

    it('shows no status at all, since a new invoice is always open', () => {
      expect(element().querySelector('#status')).toBeNull();
      expect(
        Array.from(element().querySelectorAll('.field-label')).map(text)
      ).not.toContain('Status');
    });

    it('starts as an outbound invoice, the common case', () => {
      expect(field<HTMLSelectElement>('type').value).toBe('OUT');
    });

    it('lists both directions in Portuguese', () => {
      const options = Array.from(field<HTMLSelectElement>('type').options).map(
        (option) => [option.value, text(option)]
      );

      expect(options).toEqual([
        ['OUT', 'Saída'],
        ['IN', 'Entrada']
      ]);
    });

    it('shows no error before any interaction', () => {
      expect(errors()).toEqual([]);
      expect(banner()).toBe('');
    });

    it('offers cancel going back to the listing', () => {
      expect(element().querySelector('a.btn--ghost')?.getAttribute('href')).toBe(
        '/billing/invoices'
      );
    });

    it('starts with no item and says so', () => {
      expect(rows()).toEqual([]);
      expect(text(element().querySelector('.items__empty'))).toBe(
        'Nenhum item adicionado.'
      );
    });
  });

  describe('validation', () => {
    it('blocks the submit and reveals the required field errors', async () => {
      await submit();

      expect(saved).toEqual([]);
      expect(errorOf('number')).toBe('Campo obrigatório.');
    });

    it('demands the number once the field is emptied', async () => {
      await fill('number', '6');
      await fill('number', '');

      expect(errorOf('number')).toBe('Campo obrigatório.');
    });

    it('rejects a number longer than the 30 characters the API accepts', async () => {
      await fill('number', '1000000');

      expect(errorOf('number')).toBe('O valor máximo é 999999.');
    });

    it('accepts the highest number the series allows', async () => {
      await fill('series', '1');
      await fill('number', '999999');
      await submit();

      expect(errorOf('number')).toBe('');
      expect(saved.length).toBe(1);
    });

    it('accepts a number already used by another invoice', async () => {
      await fillValidForm();
      await fill('series', '1');
      await fill('number', '1');
      await submit();

      expect(errorOf('number')).toBe('');
      expect(saved).toEqual([expect.objectContaining({ series: 1, number: 1 })]);
    });

    it('marks the invalid field visually', async () => {
      await submit();
      expect(field('number').classList).toContain('field--error');
    });

    it('clears the error as soon as the field is fixed', async () => {
      await submit();
      expect(errorOf('number')).toBe('Campo obrigatório.');

      await fill('number', '6');
      expect(errorOf('number')).toBe('');
    });
  });

  describe('emitting what was filled', () => {
    it('hands over the filled data', async () => {
      await fillValidForm();
      await submit();

      expect(saved).toEqual([{ series: 1, number: 6, type: 'OUT', items: [] }]);
    });

    it('hands over the direction that was chosen', async () => {
      await fill('series', '1');
      await fill('number', '6');
      await fill('type', 'IN');
      await submit();

      expect(saved).toEqual([expect.objectContaining({ type: 'IN' })]);
    });

    it('hands over the outbound direction when it is untouched', async () => {
      await fill('series', '1');
      await fill('number', '6');
      await submit();

      expect(saved).toEqual([expect.objectContaining({ type: 'OUT' })]);
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

  describe('items removed from the invoice', () => {
    it('emits an empty list so the API clears the items', async () => {
      await setInput('value', EXISTING);
      await removeItem(0);
      await submit();

      expect(saved).toEqual([expect.objectContaining({ items: [] })]);
    });
  });

  describe('items that are emitted', () => {
    it('hands over the snapshot of the chosen product', async () => {
      await fillValidForm();
      await addValidItem(0, 0, '2');
      await submit();

      expect(saved).toEqual([
        { series: 1, number: 6, type: 'OUT', items: [EXISTING_ITEM] }
      ]);
    });

    it('hands over the unit price that was typed over the one of the product', async () => {
      await fillValidForm();
      await addValidItem();
      await fill('item-0-unitPrice', '99.9');
      await submit();

      expect(saved).toEqual([
        expect.objectContaining({
          items: [expect.objectContaining({ unitPrice: 99.9 })]
        })
      ]);
    });

    it('hands over every row in the order they were added', async () => {
      await fillValidForm();
      await addValidItem(0, 1, '3');
      await addValidItem(1, 0, '2');
      await submit();

      expect(saved[0].items.map((item) => item.code)).toEqual([
        'PRD-0005',
        'PRD-0003'
      ]);
    });

    it('blocks the submit while a row has no product', async () => {
      await fillValidForm();
      await addItem();
      await submit();

      expect(saved).toEqual([]);
      expect(errorOf('item-0-product')).toBe('Campo obrigatório.');
    });

    it('blocks the submit on a quantity the API would reject', async () => {
      await fillValidForm();
      await addValidItem(0, 0, '0');
      await submit();

      expect(saved).toEqual([]);
      expect(errorOf('item-0-quantity')).toBe('O valor mínimo é 1.');
    });

    it('blocks the submit on a fractional quantity', async () => {
      await fillValidForm();
      await addValidItem(0, 0, '1.5');
      await submit();

      expect(saved).toEqual([]);
      expect(errorOf('item-0-quantity')).toBe('Informe um número inteiro.');
    });

    it('blocks the submit on a negative unit price', async () => {
      await fillValidForm();
      await addValidItem();
      await fill('item-0-unitPrice', '-1');
      await submit();

      expect(saved).toEqual([]);
      expect(errorOf('item-0-unitPrice')).toBe('O valor não pode ser negativo.');
    });

    it('points at the row that is wrong, not at the other one', async () => {
      await fillValidForm();
      await addValidItem(0, 0, '2');
      await addValidItem(1, 1, '0');
      await submit();

      expect(errorOf('item-0-quantity')).toBe('');
      expect(errorOf('item-1-quantity')).toBe('O valor mínimo é 1.');
    });
  });

  describe('the invoice being edited', () => {
    it('fills every field with the value it was given', async () => {
      await setInput('value', EXISTING);

      expect(field<HTMLInputElement>('series').value).toBe('1');
      expect(field<HTMLInputElement>('number').value).toBe('7');
      expect(field<HTMLSelectElement>('type').value).toBe('OUT');
      expect(element().querySelector('#status')).toBeNull();
    });

    it('fills the rows with the items it was given', async () => {
      await setInput('value', EXISTING);

      expect(rows()).toHaveLength(1);
      expect(field<HTMLSelectElement>('item-0-product').selectedIndex).toBe(1);
      expect(field<HTMLInputElement>('item-0-quantity').value).toBe('2');
      expect(field<HTMLInputElement>('item-0-unitPrice').value).toBe('150.5');
      expect(unitOf(0)).toBe('UN');
      expect(invoiceTotal()).toBe('R$ 301,00');
    });

    it('shows no error over an invoice that is still untouched', async () => {
      await setInput('value', EXISTING);

      expect(errors()).toEqual([]);
      expect(banner()).toBe('');
    });

    it('hands over the edited data', async () => {
      await setInput('value', EXISTING);
      await fill('series', '1');
      await fill('number', '42');
      await submit();

      expect(saved).toEqual([
        { series: 1, number: 42, type: 'OUT', items: [EXISTING_ITEM] }
      ]);
    });

    it('keeps the item that was already there when another is added', async () => {
      await setInput('value', EXISTING);
      await addValidItem(1, 1, '3');
      await submit();

      expect(saved[0].items).toEqual([
        EXISTING_ITEM,
        expect.objectContaining({ code: 'PRD-0005', quantity: 3 })
      ]);
    });

    it('still blocks the submit when a field was emptied', async () => {
      await setInput('value', EXISTING);
      await fill('number', '');
      await submit();

      expect(saved).toEqual([]);
      expect(errorOf('number')).toBe('Campo obrigatório.');
    });
  });

  describe('labels and states given by the page', () => {
    it('names the submit button as asked', async () => {
      await setInput('submitLabel', 'Salvar alterações');

      expect(text(submitButton())).toBe('Salvar alterações');
    });

    it('announces the save in progress', async () => {
      await setInput('saving', true);

      expect(text(submitButton())).toBe('Salvando…');
      expect(submitButton().disabled).toBe(true);
    });

    it('waits for the invoice before showing the fields', async () => {
      await setInput('loading', true);

      expect(text(element().querySelector('.form__status'))).toBe(
        'Carregando nota fiscal…'
      );
      expect(element().querySelector('#number')).toBeNull();
      expect(element().querySelector('.items')).toBeNull();
      expect(submitButton().disabled).toBe(true);
    });

    it('emits nothing while the invoice is still loading', async () => {
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
      await rejectWith({ number: 'Campo obrigatório.' });

      expect(errorOf('number')).toBe('Campo obrigatório.');
      expect(banner()).toBe('');
    });

    it('shows the server phrase exactly as it came', async () => {
      await fillValidForm();
      await rejectWith({ number: 'Já existe uma nota fiscal com esta série e número.' });

      expect(errorOf('number')).toBe('Já existe uma nota fiscal com esta série e número.');
    });

    it('shows a phrase the frontend has no copy for', async () => {
      await fillValidForm();
      await fill('type', 'OUT');
      await rejectWith({ type: 'Esta natureza de operação não está liberada.' });

      expect(errorOf('type')).toBe('Esta natureza de operação não está liberada.');
    });

    it('drops the server error as soon as the field is fixed', async () => {
      await fillValidForm();
      await rejectWith({ number: 'Campo obrigatório.' });
      expect(errorOf('number')).toBe('Campo obrigatório.');

      await fill('number', '42');
      expect(errorOf('number')).toBe('');
    });

    it('shows a general warning when no field was pointed at', async () => {
      await fillValidForm();
      await rejectWith(null, 'Não foi possível ler os dados enviados.');

      expect(banner()).toBe('Não foi possível ler os dados enviados.');
      expect(errors()).toEqual([]);
    });

    it('shows a general warning when the pointed field does not exist here', async () => {
      await fillValidForm();
      await rejectWith({ issuer: 'Emitente inválido.' }, 'Dados inválidos.');

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

    it('shows on the row the error the server pointed at', async () => {
      await fillValidForm();
      await addValidItem();
      await submit();
      await rejectWith({
        'items[0].quantity': 'O valor precisa ser maior que zero.'
      });

      expect(errorOf('item-0-quantity')).toBe('O valor precisa ser maior que zero.');
      expect(banner()).toBe('');
    });

    it('points at the row the server named, not at the first one', async () => {
      await fillValidForm();
      await addValidItem(0, 0, '2');
      await addValidItem(1, 1, '3');
      await submit();
      await rejectWith({ 'items[1].unitPrice': 'Preço acima do permitido.' });

      expect(errorOf('item-0-unitPrice')).toBe('');
      expect(errorOf('item-1-unitPrice')).toBe('Preço acima do permitido.');
    });

    it('shows a general warning when the row the server named is gone', async () => {
      await fillValidForm();
      await addValidItem();
      await submit();
      await rejectWith({ 'items[3].quantity': 'Valor inválido.' }, 'Dados inválidos.');

      expect(banner()).toBe('Dados inválidos.');
      expect(errors()).toEqual([]);
    });
  });
  describe('reporting a refused print', () => {
    const refused: Invoice = {
      ...EXISTING,
      status: 'OPEN',
      failureReason: 'INSUFFICIENT_STOCK',
      shortages: [
        { inventoryId: 3, code: 'PRD-0003', name: 'Cadeira Gamer', required: 50, available: 42 }
      ]
    };

    it('explains at the top what the stock could not cover', async () => {
      await setInput('value', refused);

      expect(text(element().querySelector('.form__warning'))).toBe(
        'Não foi possível imprimir: PRD-0003 tem 42 em estoque e a nota pede 50.'
      );
    });

    it('reports the balance beside the quantity that has to change', async () => {
      await setInput('value', refused);

      const warnings = Array.from(element().querySelectorAll('.field-warning')).map(text);

      expect(warnings).toContain('No fechamento havia 42(UN) no estoque');
    });

    it('says nothing when the invoice was never refused', async () => {
      await setInput('value', EXISTING);

      expect(element().querySelector('.form__warning')).toBeNull();
      expect(element().querySelector('.field-warning')).toBeNull();
    });
  });
});

describe('InvoiceForm loading an inbound invoice', () => {
  it('fills the direction with the one it was given', async () => {
    await TestBed.configureTestingModule({
      imports: [InvoiceForm],
      providers: [provideRouter([])]
    }).compileComponents();

    const fixture = TestBed.createComponent(InvoiceForm);
    fixture.componentRef.setInput('value', { ...EXISTING, type: 'IN' });
    await fixture.whenStable();

    const element = fixture.nativeElement as HTMLElement;

    expect(element.querySelector<HTMLSelectElement>('#type')!.value).toBe('IN');
  });

});
