import { HttpErrorResponse } from '@angular/common/http';
import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router, convertToParamMap, provideRouter } from '@angular/router';
import { Observable, Subject, of, tap, throwError } from 'rxjs';

import { CatalogProduct } from '../catalog.model';
import { CatalogService } from '../catalog.service';
import { FlashService } from '../../../shared/flash/flash.service';
import { Invoice, InvoiceItemPayload, InvoicePayload } from '../invoice.model';
import { InvoiceService } from '../invoice.service';
import { InvoiceEdit } from './invoice-edit';

const PRODUCTS: CatalogProduct[] = [
  { id: 3, code: 'PRD-0003', name: 'Cadeira Gamer', unit: 'UN', price: 150.5 },
  { id: 5, code: 'PRD-0005', name: 'Mesa de Escritório', unit: 'CX', price: 899 }
];

const EXISTING_ITEM: InvoiceItemPayload = {
  inventoryId: 3,
  code: 'PRD-0003',
  name: 'Cadeira Gamer',
  unit: 'UN',
  quantity: 2,
  unitPrice: 150.5,
  icmsRate: 18
};

const EXISTING: Invoice = {
  id: 7,
  series: 1, number: 7, formattedNumber: '001/000007',
  type: 'OUT',
  status: 'OPEN',
  total: 301,
  icmsBase: 301,
  icmsValue: 54.18,
  items: [{ ...EXISTING_ITEM, id: 1, productId: 11, total: 301, icmsBase: 301, icmsValue: 54.18 }]
};

describe('InvoiceEdit', () => {
  let fixture: ComponentFixture<InvoiceEdit>;
  let service: {
    create: ReturnType<typeof vi.fn>;
    get: ReturnType<typeof vi.fn>;
    update: ReturnType<typeof vi.fn>;
  };
  let catalogService: { products: ReturnType<typeof signal<CatalogProduct[]>>; list: ReturnType<typeof vi.fn> };
  let navigate: ReturnType<typeof vi.spyOn>;
  let flash: { error: ReturnType<typeof vi.fn>; success: ReturnType<typeof vi.fn> };

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

  const submitButton = () =>
    element().querySelector<HTMLButtonElement>('button[type="submit"]')!;

  const rows = () => Array.from(element().querySelectorAll('.items__table tbody tr'));

  const addItem = async () => {
    element().querySelector<HTMLButtonElement>('.items__header button')!.click();
    await fixture.whenStable();
  };

  const chooseProduct = async (row: number, productIndex: number) => {
    const select = field<HTMLSelectElement>(`item-${row}-product`);
    select.selectedIndex = productIndex + 1;
    select.dispatchEvent(new Event('change'));
    select.dispatchEvent(new Event('blur'));
    await fixture.whenStable();
  };

  const mount = async (
    load: () => Observable<Invoice> = () => of(EXISTING),
    id = '7',
    loadProducts: () => Observable<CatalogProduct[]> = () => of(PRODUCTS)
  ) => {
    TestBed.resetTestingModule();

    service = {
      create: vi.fn(),
      get: vi.fn(load),
      update: vi.fn((invoiceId: number, data: InvoicePayload) =>
        of({ ...data, id: invoiceId, total: 0 })
      )
    };

    catalogService = {
      products: signal<CatalogProduct[]>([]),
      list: vi.fn(() =>
        loadProducts().pipe(tap((products) => catalogService.products.set(products)))
      )
    };

    flash = { error: vi.fn(), success: vi.fn() };

    await TestBed.configureTestingModule({
      imports: [InvoiceEdit],
      providers: [
        provideRouter([]),
        { provide: InvoiceService, useValue: service },
        { provide: CatalogService, useValue: catalogService },
        { provide: FlashService, useValue: flash },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: convertToParamMap({ id }) } }
        }
      ]
    }).compileComponents();

    navigate = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);

    fixture = TestBed.createComponent(InvoiceEdit);
    await fixture.whenStable();
  };

  beforeEach(async () => {
    await mount();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('loading the invoice', () => {
    it('asks the API for the invoice in the route', () => {
      expect(service.get).toHaveBeenCalledWith(7);
    });

    it('fills every field with what came back', () => {
      expect(field<HTMLInputElement>('series').value).toBe('1');
      expect(field<HTMLInputElement>('number').value).toBe('7');
    });

    it('shows the direction without letting it be changed, and no status at all', () => {
      expect(element().querySelector('select#type')).toBeNull();
      expect(element().querySelector('#status')).toBeNull();

      expect(
        Array.from(element().querySelectorAll('.field-static')).map(text)
      ).toEqual(['Saída']);
    });

    it('announces that it is editing, not creating', () => {
      expect(text(element().querySelector('.page__title'))).toBe(
        'Editar nota fiscal'
      );
      expect(text(submitButton())).toBe('Salvar alterações');
    });

    it('shows no error over an invoice that is still untouched', () => {
      expect(errors()).toEqual([]);
      expect(failure()).toBe('');
      expect(flash.error).not.toHaveBeenCalled();
    });

    it('waits for the invoice before showing the fields', async () => {
      await mount(() => new Subject<Invoice>());

      expect(text(element().querySelector('.form__status'))).toBe(
        'Carregando nota fiscal…'
      );
      expect(element().querySelector('#number')).toBeNull();
      expect(submitButton().disabled).toBe(true);
    });
  });

  describe('saving the changes', () => {
    it('sends the edited data to the update of that id', async () => {
      await fill('series', '1');
      await fill('series', '1');
      await fill('number', '42');
      await submit();

      expect(service.update).toHaveBeenCalledWith(7, {
        series: 1, number: 42,
        items: [EXISTING_ITEM]
      });
    });

    it('never sends the type nor the status, which the edit does not own', async () => {
      await submit();

      expect(service.update).toHaveBeenCalledWith(
        7,
        expect.not.objectContaining({ type: 'OUT' })
      );
      expect(service.update).toHaveBeenCalledWith(
        7,
        expect.not.objectContaining({ status: 'OPEN' })
      );
    });

    it('never creates a second invoice', async () => {
      await submit();

      expect(service.create).not.toHaveBeenCalled();
      expect(service.update).toHaveBeenCalledTimes(1);
    });

    it('sends the number that was typed', async () => {
      await fill('series', '1');
      await fill('number', '99');
      await submit();

      expect(service.update).toHaveBeenCalledWith(
        7,
        expect.objectContaining({ series: 1, number: 99 })
      );
    });

    it('goes back to the listing after saving', async () => {
      await submit();

      expect(navigate).toHaveBeenCalledWith(['/billing/invoices']);
    });

    it('still blocks the submit when a field was emptied', async () => {
      await fill('number', '');
      await submit();

      expect(service.update).not.toHaveBeenCalled();
      expect(errorOf('number')).toBe('Campo obrigatório.');
    });

    it('shows on the field the error the server pointed at', async () => {
      service.update.mockReturnValue(
        throwError(
          () =>
            new HttpErrorResponse({
              status: 400,
              error: { errors: { number: 'Campo obrigatório.' } }
            })
        )
      );

      await submit();

      expect(errorOf('number')).toBe('Campo obrigatório.');
      expect(navigate).not.toHaveBeenCalled();
    });

    it('warns without leaking the message of a 500', async () => {
      service.update.mockReturnValue(
        throwError(
          () =>
            new HttpErrorResponse({
              status: 500,
              error: { message: 'erro ao atualizar a nota fiscal' }
            })
        )
      );

      await submit();

      expect(failure()).toBe('Não foi possível salvar a nota fiscal. Tente novamente.');
      expect(navigate).not.toHaveBeenCalled();
    });
  });

  describe('items of the invoice', () => {
    it('shows the item that came with the invoice', () => {
      expect(rows()).toHaveLength(1);
      expect(field<HTMLSelectElement>('item-0-product').selectedIndex).toBe(1);
      expect(field<HTMLInputElement>('item-0-quantity').value).toBe('2');
    });

    it('asks the billing catalog for the products it can offer', () => {
      expect(catalogService.list).toHaveBeenCalledTimes(1);
    });

    it('sends the item that was added on top of the existing one', async () => {
      await addItem();
      await chooseProduct(1, 1);
      await fill('item-1-quantity', '3');
      await submit();

      expect(service.update).toHaveBeenCalledWith(7, {
        series: 1, number: 7,
        items: [
          EXISTING_ITEM,
          {
            inventoryId: 5,
            code: 'PRD-0005',
            name: 'Mesa de Escritório',
            unit: 'CX',
            quantity: 3,
            unitPrice: 899,
            icmsRate: 0
          }
        ]
      });
    });

    it('sends an empty list when every item was removed', async () => {
      element().querySelector<HTMLButtonElement>('.items__remove')!.click();
      await fixture.whenStable();
      await submit();

      expect(service.update).toHaveBeenCalledWith(
        7,
        expect.objectContaining({ items: [] })
      );
    });

    it('shows on the row the error the server pointed at', async () => {
      service.update.mockReturnValue(
        throwError(
          () =>
            new HttpErrorResponse({
              status: 400,
              error: { errors: { 'items[0].quantity': 'O valor precisa ser maior que zero.' } }
            })
        )
      );

      await submit();

      expect(errorOf('item-0-quantity')).toBe('O valor precisa ser maior que zero.');
      expect(navigate).not.toHaveBeenCalled();
    });

    it('warns when the inventory cannot be reached', async () => {
      await mount(undefined, '7', () =>
        throwError(() => new HttpErrorResponse({ status: 500 }))
      );

      expect(flash.error).toHaveBeenCalledWith(
        'Não foi possível carregar os produtos. Tente novamente.'
      );
      expect(text(element().querySelector('.items__note'))).toBe(
        'Não foi possível carregar os produtos do estoque.'
      );
    });
  });

  describe('invoice that cannot be loaded', () => {
    const failLoadWith = (status: number) =>
      mount(() => throwError(() => new HttpErrorResponse({ status })));

    it('goes back to the listing when the invoice does not exist', async () => {
      await failLoadWith(404);

      expect(navigate).toHaveBeenCalledWith(['/billing/invoices']);
    });

    it('goes back to the listing when the API is down', async () => {
      await failLoadWith(500);

      expect(navigate).toHaveBeenCalledWith(['/billing/invoices']);
    });

    it('says on the way out that the invoice does not exist', async () => {
      await failLoadWith(404);

      expect(flash.error).toHaveBeenCalledWith('Nota fiscal não encontrada.');
    });

    it('blames the API when the failure was not a 404', async () => {
      await failLoadWith(500);

      expect(flash.error).toHaveBeenCalledWith(
        'Não foi possível carregar a nota fiscal. Tente novamente.'
      );
    });

    it('warns once per failed load', async () => {
      await failLoadWith(404);

      expect(flash.error).toHaveBeenCalledTimes(1);
    });

    it('never shows the empty form while it leaves the screen', async () => {
      await failLoadWith(404);

      expect(element().querySelector('#number')).toBeNull();
      expect(text(element().querySelector('.form__status'))).toBe(
        'Carregando nota fiscal…'
      );
    });

    it('saves nothing on the way out', async () => {
      await failLoadWith(404);

      await submit();

      expect(service.update).not.toHaveBeenCalled();
      expect(service.create).not.toHaveBeenCalled();
    });
  });

  describe('a closed invoice reached by its url', () => {
    beforeEach(async () => {
      await mount(() => of({ ...EXISTING, status: 'CLOSED' as const }));
    });

    it('never shows the fields', () => {
      expect(element().querySelector('#number')).toBeNull();
      expect(element().querySelector('#type')).toBeNull();
    });

    it('says why and goes back to the listing', () => {
      expect(flash.error).toHaveBeenCalledWith(
        'Notas fiscais fechadas não podem ser alteradas.'
      );
      expect(navigate).toHaveBeenCalledWith(['/billing/invoices']);
    });

    it('never saves anything', () => {
      expect(service.update).not.toHaveBeenCalled();
    });
  });
});
