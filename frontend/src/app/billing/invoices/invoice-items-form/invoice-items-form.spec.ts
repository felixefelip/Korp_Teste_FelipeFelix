import { registerLocaleData } from '@angular/common';
import localePt from '@angular/common/locales/pt';
import { LOCALE_ID } from '@angular/core';
import { FormArray } from '@angular/forms';
import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Product } from '../../../inventory/products/product.model';
import { InvoiceItem } from '../invoice.model';
import { InvoiceItemsForm, ItemGroup, newItemArray } from './invoice-items-form';

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

const EXISTING_ITEM: InvoiceItem = {
  id: 1,
  productId: 11,
  inventoryId: 3,
  code: 'PRD-0003',
  name: 'Cadeira Gamer',
  unit: 'UN',
  quantity: 2,
  unitPrice: 150.5,
  total: 301,
  icmsRate: 18,
  icmsBase: 301,
  icmsValue: 54.18
};

describe('InvoiceItemsForm', () => {
  let fixture: ComponentFixture<InvoiceItemsForm>;
  let items: FormArray<ItemGroup>;

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
    const input = field<HTMLInputElement>(id);
    input.value = value;
    input.dispatchEvent(new Event('input'));
    input.dispatchEvent(new Event('blur'));
    await fixture.whenStable();
  };

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

  const rows = () => Array.from(element().querySelectorAll('tbody tr'));

  const rowTotals = () =>
    Array.from(element().querySelectorAll('.items__row-total')).map(text);

  const total = () => text(element().querySelector('.items__total'));

  const rowICMS = () =>
    Array.from(element().querySelectorAll('.items__row-icms')).map(text);

  const totalICMS = () => text(element().querySelector('.items__total-icms'));

  const unitOf = (row: number) => text(rows()[row].querySelector('.items__unit'));

  const errorOf = (id: string) =>
    text(field(id).parentElement?.querySelector('.field-error'));

  const errors = () =>
    Array.from(element().querySelectorAll('.field-error')).map(text);

  const empty = () => text(element().querySelector('.items__empty'));

  const mount = async (given: FormArray<ItemGroup> = newItemArray()) => {
    TestBed.resetTestingModule();

    await TestBed.configureTestingModule({
      imports: [InvoiceItemsForm],
      providers: [{ provide: LOCALE_ID, useValue: 'pt-BR' }]
    }).compileComponents();

    items = given;

    fixture = TestBed.createComponent(InvoiceItemsForm);
    fixture.componentRef.setInput('items', items);
    fixture.componentRef.setInput('products', PRODUCTS);
    await fixture.whenStable();
  };

  const addValidItem = async (row = 0, productIndex = 0, quantity = '2') => {
    await addItem();
    await chooseProduct(row, productIndex);
    await fill(`item-${row}-quantity`, quantity);
  };

  beforeEach(async () => {
    await mount();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('with no item yet', () => {
    it('shows the empty state instead of the table', () => {
      expect(element().querySelector('table')).toBeNull();
      expect(empty()).toBe('Nenhum item adicionado.');
    });

    it('counts the items it holds', () => {
      expect(text(element().querySelector('.items__subtitle'))).toBe(
        '0 item(ns) na nota fiscal'
      );
    });

    it('asks for a product to be registered when the catalog is empty', async () => {
      await setInput('products', []);

      expect(addButton().disabled).toBe(true);
      expect(empty()).toBe('Cadastre um produto no estoque antes de adicionar itens.');
      expect(element().querySelector('.items__note')).toBeNull();
    });

    it('blames the inventory when the catalog could not be loaded', async () => {
      await setInput('products', []);
      await setInput('productsFailed', true);

      expect(addButton().disabled).toBe(true);
      expect(text(element().querySelector('.items__note'))).toBe(
        'Não foi possível carregar os produtos do estoque.'
      );
      expect(empty()).toBe('Nenhum item adicionado.');
    });
  });

  describe('adding a row', () => {
    it('pushes the row into the array it was given', async () => {
      await addItem();

      expect(items.length).toBe(1);
      expect(rows()).toHaveLength(1);
    });

    it('offers every product of the catalog', async () => {
      await addItem();

      const options = Array.from(
        field<HTMLSelectElement>('item-0-product').options
      ).map(text);

      expect(options).toEqual([
        'Selecione o produto',
        'PRD-0003 · Cadeira Gamer',
        'PRD-0005 · Mesa de Escritório'
      ]);
    });

    it('starts the row with one unit and no product chosen', async () => {
      await addItem();

      expect(field<HTMLSelectElement>('item-0-product').selectedIndex).toBe(0);
      expect(field<HTMLInputElement>('item-0-quantity').value).toBe('1');
      expect(field<HTMLInputElement>('item-0-unitPrice').value).toBe('');
    });

    it('takes unit and price from the product that was chosen', async () => {
      await addItem();
      await chooseProduct(0, 1);

      expect(unitOf(0)).toBe('CX');
      expect(field<HTMLInputElement>('item-0-unitPrice').value).toBe('899');
      expect(items.at(0).getRawValue()).toMatchObject({
        inventoryId: 5,
        code: 'PRD-0005',
        name: 'Mesa de Escritório',
        unit: 'CX'
      });
    });

    it('replaces unit and price when the product changes', async () => {
      await addItem();
      await chooseProduct(0, 1);
      await chooseProduct(0, 0);

      expect(unitOf(0)).toBe('UN');
      expect(field<HTMLInputElement>('item-0-unitPrice').value).toBe('150.5');
    });

    it('keeps a unit price typed over the one of the product', async () => {
      await addValidItem();
      await fill('item-0-unitPrice', '99.9');

      expect(items.at(0).getRawValue().unitPrice).toBe(99.9);
    });

    it('adds a second row without touching the first', async () => {
      await addValidItem(0, 0, '2');
      await addValidItem(1, 1, '3');

      expect(rows()).toHaveLength(2);
      expect(unitOf(0)).toBe('UN');
      expect(unitOf(1)).toBe('CX');
      expect(field<HTMLInputElement>('item-0-quantity').value).toBe('2');
      expect(field<HTMLInputElement>('item-1-quantity').value).toBe('3');
    });
  });

  describe('totals', () => {
    it('multiplies quantity by unit price on the row', async () => {
      await addValidItem(0, 0, '2');

      expect(rowTotals()).toEqual(['R$ 301,00']);
    });

    it('sums every row into the invoice total', async () => {
      await addValidItem(0, 0, '2');
      await addValidItem(1, 1, '3');

      expect(rowTotals()).toEqual(['R$ 301,00', 'R$ 2.697,00']);
      expect(total()).toBe('R$ 2.998,00');
    });

    it('follows the quantity as it is typed', async () => {
      await addValidItem(0, 0, '2');
      await fill('item-0-quantity', '4');

      expect(rowTotals()).toEqual(['R$ 602,00']);
      expect(total()).toBe('R$ 602,00');
    });

    it('shows zero for a row with no product yet', async () => {
      await addItem();

      expect(rowTotals()).toEqual(['R$ 0,00']);
      expect(total()).toBe('R$ 0,00');
    });
  });

  describe('icms', () => {
    it('starts a row untaxed', async () => {
      await addItem();

      expect(field<HTMLInputElement>('item-0-icmsRate').value).toBe('0');
      expect(rowICMS()).toEqual(['R$ 0,00']);
      expect(totalICMS()).toBe('R$ 0,00');
    });

    it('applies the typed rate over the total of the row', async () => {
      await addValidItem(0, 0, '2');
      await fill('item-0-icmsRate', '18');

      expect(rowTotals()).toEqual(['R$ 301,00']);
      expect(rowICMS()).toEqual(['R$ 54,18']);
    });

    it('sums the icms of every row without touching the total of the invoice', async () => {
      await addValidItem(0, 0, '2');
      await fill('item-0-icmsRate', '18');
      await addValidItem(1, 1, '3');
      await fill('item-1-icmsRate', '12');

      expect(totalICMS()).toBe('R$ 377,82');
      expect(total()).toBe('R$ 2.998,00');
    });

    it('taxes each row by its own rate', async () => {
      await addValidItem(0, 0, '2');
      await fill('item-0-icmsRate', '18');
      await addValidItem(1, 1, '3');

      expect(rowICMS()).toEqual(['R$ 54,18', 'R$ 0,00']);
    });

    it('shows the rate a row came with', async () => {
      await mount(newItemArray([EXISTING_ITEM]));

      expect(field<HTMLInputElement>('item-0-icmsRate').value).toBe('18');
      expect(rowICMS()).toEqual(['R$ 54,18']);
    });

    it('rejects a rate above one hundred percent', async () => {
      await addValidItem();
      await fill('item-0-icmsRate', '101');
      await setInput('submitted', true);

      expect(errorOf('item-0-icmsRate')).toBe('O valor máximo é 100.');
    });

    it('rejects a negative rate', async () => {
      await addValidItem();
      await fill('item-0-icmsRate', '-1');
      await setInput('submitted', true);

      expect(errorOf('item-0-icmsRate')).toBe('O valor não pode ser negativo.');
    });
  });

  describe('removing a row', () => {
    it('drops it from the array it was given', async () => {
      await addValidItem(0, 0, '2');
      await addValidItem(1, 1, '3');
      await removeItem(0);

      expect(items.length).toBe(1);
      expect(rows()).toHaveLength(1);
    });

    it('keeps on screen the row that was not removed', async () => {
      await addValidItem(0, 0, '2');
      await addValidItem(1, 1, '3');
      await removeItem(0);

      expect(unitOf(0)).toBe('CX');
      expect(field<HTMLInputElement>('item-0-quantity').value).toBe('3');
    });

    it('goes back to the empty state when the last row leaves', async () => {
      await addValidItem();
      await removeItem(0);

      expect(items.length).toBe(0);
      expect(element().querySelector('table')).toBeNull();
      expect(empty()).toBe('Nenhum item adicionado.');
    });
  });

  describe('rows that came ready', () => {
    beforeEach(async () => {
      await mount(newItemArray([EXISTING_ITEM]));
    });

    it('shows one row per item of the array', () => {
      expect(rows()).toHaveLength(1);
      expect(text(element().querySelector('.items__subtitle'))).toBe(
        '1 item(ns) na nota fiscal'
      );
    });

    it('fills the row with what the item carried', () => {
      expect(field<HTMLSelectElement>('item-0-product').selectedIndex).toBe(1);
      expect(field<HTMLInputElement>('item-0-quantity').value).toBe('2');
      expect(field<HTMLInputElement>('item-0-unitPrice').value).toBe('150.5');
      expect(unitOf(0)).toBe('UN');
      expect(total()).toBe('R$ 301,00');
    });

    it('shows no error over a row nobody touched', () => {
      expect(errors()).toEqual([]);
    });
  });

  describe('errors of the row', () => {
    it('stays quiet until the field is touched', async () => {
      await addItem();
      await fill('item-0-quantity', '0');
      items.at(0).controls.quantity.markAsUntouched();
      await fixture.whenStable();

      expect(errorOf('item-0-quantity')).toBe('');
    });

    it('speaks once the page says it was submitted', async () => {
      await addItem();
      await fill('item-0-quantity', '0');
      await setInput('submitted', true);

      expect(errorOf('item-0-quantity')).toBe('O valor mínimo é 1.');
    });

    it('demands a product on a row that has none', async () => {
      await addItem();
      await setInput('submitted', true);

      expect(errorOf('item-0-product')).toBe('Campo obrigatório.');
    });

    it('rejects a fractional quantity', async () => {
      await addValidItem(0, 0, '1.5');
      await setInput('submitted', true);

      expect(errorOf('item-0-quantity')).toBe('Informe um número inteiro.');
    });

    it('rejects a negative unit price', async () => {
      await addValidItem();
      await fill('item-0-unitPrice', '-1');
      await setInput('submitted', true);

      expect(errorOf('item-0-unitPrice')).toBe('O valor não pode ser negativo.');
    });

    it('shows the message the server wrote on the control', async () => {
      await addValidItem();
      await setInput('submitted', true);
      items.at(0).controls.quantity.setErrors({
        server: 'O valor precisa ser maior que zero.'
      });
      await fixture.whenStable();

      expect(errorOf('item-0-quantity')).toBe('O valor precisa ser maior que zero.');
    });

    it('points at the row that is wrong, not at the other one', async () => {
      await addValidItem(0, 0, '2');
      await addValidItem(1, 1, '0');
      await setInput('submitted', true);

      expect(errorOf('item-0-quantity')).toBe('');
      expect(errorOf('item-1-quantity')).toBe('O valor mínimo é 1.');
    });
  });
});
