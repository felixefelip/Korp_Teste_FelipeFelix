import { HttpErrorResponse } from '@angular/common/http';
import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router, convertToParamMap, provideRouter } from '@angular/router';
import { Observable, of, throwError } from 'rxjs';

import { FlashService } from '../../../shared/flash/flash.service';
import { Product } from '../../products/product.model';
import { ProductService } from '../../products/product.service';
import { Movement } from '../movement.model';
import { MovementService } from '../movement.service';
import { MovementList } from './movement-list';

const PRODUCT: Product = {
  id: 7,
  code: 'PRD-0007',
  name: 'Cadeira de escritório',
  unit: 'CX',
  price: 750.5,
  stock: 10
};

const MOVEMENTS: Movement[] = [
  {
    id: 3,
    productId: 7,
    type: 'out',
    origin: 'invoice',
    quantity: 4,
    confirmed: false,
    billingInvoiceItemId: 33,
    billingInvoiceId: 42
  },
  {
    id: 2,
    productId: 7,
    type: 'out',
    origin: 'adjustment',
    quantity: 2,
    confirmed: true,
    billingInvoiceItemId: null,
    billingInvoiceId: null
  },
  {
    id: 1,
    productId: 7,
    type: 'in',
    origin: 'adjustment',
    quantity: 12,
    confirmed: true,
    billingInvoiceItemId: null,
    billingInvoiceId: null
  }
];

describe('MovementList', () => {
  let fixture: ComponentFixture<MovementList>;
  let movements: ReturnType<typeof signal<Movement[]>>;
  let flash: { error: ReturnType<typeof vi.fn>; success: ReturnType<typeof vi.fn> };
  let navigate: ReturnType<typeof vi.spyOn>;

  const element = () => fixture.nativeElement as HTMLElement;

  const text = (el: Element | null | undefined) =>
    (el?.textContent ?? '').replace(/[  ]/g, ' ').trim();

  const rows = () =>
    Array.from(element().querySelectorAll('tbody tr')).filter(
      (row) => !row.querySelector('.table__empty')
    );

  const cells = (row: Element) => Array.from(row.querySelectorAll('td')).map(text);

  const balance = () =>
    Array.from(element().querySelectorAll('.balance__value')).map(text);

  const empty = () => text(element().querySelector('.table__empty'));

  const mount = async (
    load: () => Observable<Movement[]> = () => of(MOVEMENTS),
    loadProduct: () => Observable<Product> = () => of(PRODUCT)
  ) => {
    TestBed.resetTestingModule();

    movements = signal<Movement[]>([]);

    const movementService = {
      movements: movements.asReadonly(),
      list: vi.fn(() =>
        new Observable<Movement[]>((subscriber) =>
          load().subscribe({
            next: (value) => {
              movements.set(value);
              subscriber.next(value);
              subscriber.complete();
            },
            error: (error) => subscriber.error(error)
          })
        )
      ),
      get: vi.fn(),
      create: vi.fn(),
      update: vi.fn()
    };

    flash = { error: vi.fn(), success: vi.fn() };

    await TestBed.configureTestingModule({
      imports: [MovementList],
      providers: [
        provideRouter([]),
        { provide: MovementService, useValue: movementService },
        {
          provide: ProductService,
          useValue: { get: vi.fn(loadProduct), list: vi.fn(), products: signal([]) }
        },
        { provide: FlashService, useValue: flash },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: convertToParamMap({ id: '7' }) } }
        }
      ]
    }).compileComponents();

    navigate = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);

    fixture = TestBed.createComponent(MovementList);
    await fixture.whenStable();
  };

  beforeEach(async () => {
    await mount();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('listing', () => {
    it('shows every movement of the product', () => {
      expect(rows()).toHaveLength(3);
    });

    it('shows the movement columns in the expected order', () => {
      expect(cells(rows()[2])).toEqual([
        'Entrada',
        'Ajuste',
        '12',
        'Confirmada',
        'Editar'
      ]);
    });

    it('translates the type and the origin', () => {
      expect(cells(rows()[0]).slice(0, 2)).toEqual(['Saída', 'Nota fiscal']);
    });

    it('names the product it belongs to', () => {
      expect(text(element().querySelector('.page__subtitle'))).toBe(
        'PRD-0007 · Cadeira de escritório'
      );
    });

    it('offers no edit for a movement born from an invoice', () => {
      const row = rows()[0];

      expect(row.querySelector('a.table__action')).toBeNull();
      expect(text(row.querySelector('.table__action--off'))).toBe('Gerada por nota');
    });

    it('links the editable ones to their own form', () => {
      expect(
        rows()[1].querySelector<HTMLAnchorElement>('a.table__action')!.getAttribute('href')
      ).toBe('/inventory/products/7/movements/2/edit');
    });

    it('says so when the ledger is empty', async () => {
      await mount(() => of([]));

      expect(rows()).toEqual([]);
      expect(empty()).toBe('Nenhuma movimentação registrada para este produto.');
    });
  });

  describe('the balances', () => {
    it('shows what is on the shelf, what is held and what is left', () => {
      expect(balance()).toEqual(['10', '4', '6']);
    });

    it('counts only the unconfirmed exits as reserved', async () => {
      await mount(() =>
        of([
          { ...MOVEMENTS[0], confirmed: true },
          { ...MOVEMENTS[2], confirmed: false }
        ])
      );

      expect(balance()).toEqual(['10', '0', '10']);
    });
  });

  describe('when the API fails', () => {
    it('offers to try again', async () => {
      await mount(() => throwError(() => new HttpErrorResponse({ status: 500 })));

      expect(text(element().querySelector('.table__error'))).toContain(
        'Não foi possível carregar as movimentações.'
      );
    });

    it('offers to try again when only the ledger is gone, without leaving', async () => {
      await mount(() => throwError(() => new HttpErrorResponse({ status: 404 })));

      expect(text(element().querySelector('.table__error'))).toContain(
        'Não foi possível carregar as movimentações.'
      );
      expect(flash.error).not.toHaveBeenCalled();
      expect(navigate).not.toHaveBeenCalled();
    });

    it('goes back to the products when the product is gone', async () => {
      await mount(
        () => of([]),
        () => throwError(() => new HttpErrorResponse({ status: 404 }))
      );

      expect(flash.error).toHaveBeenCalledWith('Produto não encontrado.');
      expect(navigate).toHaveBeenCalledWith(['/inventory/products']);
    });
  });
});
