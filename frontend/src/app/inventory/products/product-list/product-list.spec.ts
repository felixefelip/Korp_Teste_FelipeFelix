import { registerLocaleData } from '@angular/common';
import localePt from '@angular/common/locales/pt';
import { LOCALE_ID, signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { Observable, of, tap, throwError } from 'rxjs';

import { Product } from '../product.model';
import { ProductService } from '../product.service';
import { ProductList } from './product-list';

registerLocaleData(localePt, 'pt-BR');

const PRODUCTS: Product[] = [
  {
    id: 1,
    code: 'PRD-0001',
    name: 'Notebook Dell Inspiron 15',
    unit: 'UN',
    price: 4299.9,
    stock: 12
  },
  {
    id: 2,
    code: 'PRD-0002',
    name: 'Monitor LG 24" Full HD',
    unit: 'UN',
    price: 899,
    stock: 34
  },
  {
    id: 3,
    code: 'ABC-9999',
    name: 'Papel Sulfite A4',
    unit: 'CX',
    price: 27.4,
    stock: 0
  }
];

describe('ProductList', () => {
  let fixture: ComponentFixture<ProductList>;

  const text = (el: Element | null | undefined) =>
    (el?.textContent ?? '').replace(/[\u00A0\u202F]/g, ' ').trim();

  const element = () => fixture.nativeElement as HTMLElement;

  const rows = () =>
    Array.from(element().querySelectorAll('tbody tr')).filter(
      (row) => !row.querySelector('.table__empty')
    );

  const cells = (row: Element) => Array.from(row.querySelectorAll('td')).map(text);

  const names = () => rows().map((row) => cells(row)[1]);

  const openActionsOf = async (index: number) => {
    rows()[index].querySelector<HTMLButtonElement>('.menu-button__toggle')!.click();
    await fixture.whenStable();
  };

  const menuLinks = () =>
    Array.from(element().querySelectorAll<HTMLAnchorElement>('.menu-button__item')).map(
      (item) => [text(item), item.getAttribute('href')]
    );

  const typeInFilter = async (term: string) => {
    const field = element().querySelector<HTMLInputElement>('input[type="search"]')!;
    field.value = term;
    field.dispatchEvent(new Event('input'));
    await fixture.whenStable();
  };

  const mount = async (response: () => Observable<Product[]>) => {
    const products = signal<Product[]>([]);

    const service = {
      products,
      list: vi.fn(() => response().pipe(tap((list) => products.set(list))))
    };

    await TestBed.configureTestingModule({
      imports: [ProductList],
      providers: [
        provideRouter([]),
        { provide: ProductService, useValue: service },
        { provide: LOCALE_ID, useValue: 'pt-BR' }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(ProductList);
    await fixture.whenStable();

    return service;
  };

  beforeEach(async () => {
    await mount(() => of(PRODUCTS));
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('listing', () => {
    it('shows every product coming from the API', () => {
      expect(names()).toEqual([
        'Notebook Dell Inspiron 15',
        'Monitor LG 24" Full HD',
        'Papel Sulfite A4'
      ]);
    });

    it('shows the product columns in the expected order', () => {
      expect(cells(rows()[0])).toEqual([
        'PRD-0001',
        'Notebook Dell Inspiron 15',
        'UN',
        'R$ 4.299,90',
        '12',
        'Editar'
      ]);
    });

    it('formats the unit price in reais with two decimals', () => {
      const prices = rows().map((row) => cells(row)[3]);
      expect(prices).toEqual(['R$ 4.299,90', 'R$ 899,00', 'R$ 27,40']);
    });

    it('reports the product count in the subtitle', () => {
      expect(text(element().querySelector('.page__subtitle'))).toBe(
        '3 produto(s) encontrado(s)'
      );
    });

    it('offers the edit of each product without opening anything', () => {
      const edits = rows().map((row) => [
        text(row.querySelector('.menu-button__action')),
        row.querySelector('.menu-button__action')!.getAttribute('href')
      ]);

      expect(edits).toEqual([
        ['Editar', '/inventory/products/1/edit'],
        ['Editar', '/inventory/products/2/edit'],
        ['Editar', '/inventory/products/3/edit']
      ]);
    });

    it('keeps the extra actions of every row closed until one is opened', () => {
      expect(element().querySelectorAll('.menu-button__toggle')).toHaveLength(3);
      expect(element().querySelector('.menu-button__menu')).toBeNull();
    });

    it('opens the extra actions of the row that was clicked', async () => {
      await openActionsOf(0);

      expect(menuLinks()).toEqual([
        ['Movimentações', '/inventory/products/1/movements']
      ]);
    });

    it('points each row at its own product', async () => {
      await openActionsOf(2);

      expect(menuLinks()).toEqual([
        ['Movimentações', '/inventory/products/3/movements']
      ]);
    });

    it('offers the create action pointing to the form', () => {
      const action = element().querySelector('a.btn--primary');

      expect(text(action)).toBe('+ Cadastrar produto');
      expect(action?.getAttribute('href')).toBe('/inventory/products/new');
    });
  });

  describe('loading', () => {
    it('fetches the listing when the screen opens', () => {
      expect(TestBed.inject(ProductService).list).toHaveBeenCalled();
    });

    it('warns when the API does not answer', async () => {
      TestBed.resetTestingModule();
      await mount(() => throwError(() => new Error('network down')));

      expect(text(element().querySelector('.table__error'))).toContain(
        'Não foi possível carregar os produtos.'
      );
      expect(rows()).toHaveLength(0);
    });

    it('retries when the user asks for it', async () => {
      TestBed.resetTestingModule();

      let shouldFail = true;
      await mount(() =>
        shouldFail ? throwError(() => new Error('network down')) : of(PRODUCTS)
      );

      shouldFail = false;
      element().querySelector<HTMLButtonElement>('.table__error button')!.click();
      await fixture.whenStable();

      expect(names()).toHaveLength(3);
    });
  });

  describe('filter', () => {
    it('filters by name', async () => {
      await typeInFilter('monitor');
      expect(names()).toEqual(['Monitor LG 24" Full HD']);
    });

    it('filters by code', async () => {
      await typeInFilter('ABC-9999');
      expect(names()).toEqual(['Papel Sulfite A4']);
    });

    it('ignores case differences', async () => {
      await typeInFilter('nOtEbOoK');
      expect(names()).toEqual(['Notebook Dell Inspiron 15']);
    });

    it('ignores spaces around the term', async () => {
      await typeInFilter('   papel   ');
      expect(names()).toEqual(['Papel Sulfite A4']);
    });

    it('matches fragments in the middle of the name', async () => {
      await typeInFilter('dell');
      expect(names()).toEqual(['Notebook Dell Inspiron 15']);
    });

    it('returns several products when the term is common', async () => {
      await typeInFilter('PRD-');
      expect(names()).toEqual([
        'Notebook Dell Inspiron 15',
        'Monitor LG 24" Full HD'
      ]);
    });

    it('updates the subtitle count while filtering', async () => {
      await typeInFilter('PRD-');
      expect(text(element().querySelector('.page__subtitle'))).toBe(
        '2 produto(s) encontrado(s)'
      );
    });

    it('shows the empty state when nothing matches', async () => {
      await typeInFilter('inexistente');

      expect(rows()).toHaveLength(0);
      expect(text(element().querySelector('.table__empty'))).toBe(
        'Nenhum produto encontrado.'
      );
    });

    it('lists everything again when the filter is cleared', async () => {
      await typeInFilter('monitor');
      await typeInFilter('');

      expect(names()).toHaveLength(3);
    });
  });
});
