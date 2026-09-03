import { HttpErrorResponse } from '@angular/common/http';
import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';
import { Observable, Subject, of, tap, throwError } from 'rxjs';

import { CatalogProduct } from '../catalog.model';
import { CatalogService } from '../catalog.service';
import { FlashService } from '../../../shared/flash/flash.service';
import {
  Invoice,
  InvoiceDocument,
  InvoiceDraft,
  InvoicePayload
} from '../invoice.model';
import { InvoiceService } from '../invoice.service';
import { InvoiceNew } from './invoice-new';

const PRODUCTS: CatalogProduct[] = [
  { id: 3, code: 'PRD-0003', name: 'Cadeira Gamer', unit: 'UN', price: 150.5 }
];

const DOCUMENT: InvoiceDocument = { series: 1, number: 7 };

const DRAFT: InvoiceDraft = {
  type: 'OUT',
  items: [
    {
      inventoryId: 3,
      code: 'PRD-0003',
      name: 'Cadeira Gamer',
      unit: 'UN',
      quantity: 2,
      unitPrice: 150.5,
      icmsRate: 0,
      ipiRate: 0
    }
  ],
  unresolved: []
};

describe('InvoiceNew', () => {
  let fixture: ComponentFixture<InvoiceNew>;
  let service: {
    create: ReturnType<typeof vi.fn>;
    get: ReturnType<typeof vi.fn>;
    update: ReturnType<typeof vi.fn>;
    draft: ReturnType<typeof vi.fn>;
    nextDocument: ReturnType<typeof vi.fn>;
  };
  let catalogService: { products: ReturnType<typeof signal<CatalogProduct[]>>; list: ReturnType<typeof vi.fn> };
  let flash: { error: ReturnType<typeof vi.fn>; success: ReturnType<typeof vi.fn> };
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


  const promptField = () =>
    element().querySelector<HTMLTextAreaElement>('.prompt__field')!;

  const promptButton = () =>
    element().querySelector<HTMLButtonElement>('.prompt__actions .btn')!;

  const promptFailure = () => text(element().querySelector('.prompt .form__failure'));

  const issues = () =>
    Array.from(element().querySelectorAll('.prompt__issue')).map(text);

  const rows = () => element().querySelectorAll('.items__table tbody tr').length;

  const describePrompt = async (value: string) => {
    const field = promptField();
    field.value = value;
    field.dispatchEvent(new Event('input'));
    await fixture.whenStable();
  };

  const generate = async (value: string) => {
    await describePrompt(value);
    promptButton().click();
    await fixture.whenStable();
  };

  const rejectDraftWith = (body: unknown, status: number) =>
    service.draft.mockReturnValue(
      throwError(() => new HttpErrorResponse({ status, error: body }))
    );

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

  const fillValidForm = async () => {
    await fill('series', '1');
    await fill('number', '6');
  };

  const mount = async (
    loadProducts: () => Observable<CatalogProduct[]> = () => of(PRODUCTS),
    loadDocument: () => Observable<InvoiceDocument> = () => of(DOCUMENT)
  ) => {
    TestBed.resetTestingModule();

    service = {
      create: vi.fn((data: InvoicePayload) => of({ ...data, id: 6, total: 0 })),
      get: vi.fn(),
      update: vi.fn(),
      draft: vi.fn(() => of(DRAFT)),
      nextDocument: vi.fn(() => loadDocument())
    };

    catalogService = {
      products: signal<CatalogProduct[]>([]),
      list: vi.fn(() =>
        loadProducts().pipe(tap((products) => catalogService.products.set(products)))
      )
    };

    flash = { error: vi.fn(), success: vi.fn() };

    await TestBed.configureTestingModule({
      imports: [InvoiceNew],
      providers: [
        provideRouter([]),
        { provide: InvoiceService, useValue: service },
        { provide: CatalogService, useValue: catalogService },
        { provide: FlashService, useValue: flash }
      ]
    }).compileComponents();

    navigate = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);

    fixture = TestBed.createComponent(InvoiceNew);
    await fixture.whenStable();
  };

  beforeEach(async () => {
    await mount();
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
        series: 1, number: 6,
        type: 'OUT',
        items: []
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
      await fill('number', '');
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

  describe('products of the catalog', () => {
    it('asks the billing catalog for the products it can offer', () => {
      expect(catalogService.list).toHaveBeenCalledTimes(1);
    });

    it('offers in the item row the products that came back', async () => {
      await addItem();

      const options = Array.from(
        field<HTMLSelectElement>('item-0-product').options
      ).map(text);

      expect(options).toEqual(['Selecione o produto', 'PRD-0003 · Cadeira Gamer']);
    });

    it('sends the item with the snapshot of the chosen product', async () => {
      await fillValidForm();
      await addItem();
      await chooseProduct(0, 0);
      await fill('item-0-quantity', '2');
      await submit();

      expect(service.create).toHaveBeenCalledWith({
        series: 1, number: 6,
        type: 'OUT',
        items: [
          {
            inventoryId: 3,
            code: 'PRD-0003',
            name: 'Cadeira Gamer',
            unit: 'UN',
            quantity: 2,
            unitPrice: 150.5,
            icmsRate: 0,
            ipiRate: 0
          }
        ]
      });
    });

    it('warns when the inventory cannot be reached', async () => {
      await mount(() => throwError(() => new HttpErrorResponse({ status: 500 })));

      expect(flash.error).toHaveBeenCalledWith(
        'Não foi possível carregar os produtos. Tente novamente.'
      );
    });

    it('says on the form that the products are the ones missing', async () => {
      await mount(() => throwError(() => new HttpErrorResponse({ status: 500 })));

      expect(text(element().querySelector('.items__note'))).toBe(
        'Não foi possível carregar os produtos do estoque.'
      );
    });

    it('still lets the invoice be saved without items', async () => {
      await mount(() => throwError(() => new HttpErrorResponse({ status: 500 })));

      await fillValidForm();
      await submit();

      expect(service.create).toHaveBeenCalledTimes(1);
      expect(navigate).toHaveBeenCalledWith(['/billing/invoices']);
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
      await fill('number', '42');
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

  describe('drafting with AI', () => {
    it('sends the typed prompt to the service', async () => {
      await generate('vender 2 cadeiras gamer');

      expect(service.draft).toHaveBeenCalledWith('vender 2 cadeiras gamer');
    });

    it('trims the prompt before sending it', async () => {
      await generate('   vender 2 cadeiras gamer   ');

      expect(service.draft).toHaveBeenCalledWith('vender 2 cadeiras gamer');
    });

    it('never asks for a draft of an empty prompt', async () => {
      await describePrompt('   ');
      promptButton().click();
      await fixture.whenStable();

      expect(service.draft).not.toHaveBeenCalled();
    });

    it('fills the item rows with what came back', async () => {
      expect(rows()).toBe(0);

      await generate('vender 2 cadeiras gamer');

      expect(rows()).toBe(1);
      expect(field<HTMLInputElement>('item-0-quantity').value).toBe('2');
    });

    it('saves the drafted items through the usual path', async () => {
      await fillValidForm();
      await generate('vender 2 cadeiras gamer');
      await submit();

      expect(service.create).toHaveBeenCalledWith({
        series: 1,
        number: 6,
        type: 'OUT',
        items: DRAFT.items
      });
    });

    it('keeps the series and number already typed', async () => {
      await fillValidForm();
      await generate('vender 2 cadeiras gamer');

      expect(field<HTMLInputElement>('series').value).toBe('1');
      expect(field<HTMLInputElement>('number').value).toBe('6');
    });

    it('switches the type when the draft describes an entry', async () => {
      service.draft.mockReturnValue(of({ ...DRAFT, type: 'IN' as const }));

      await fillValidForm();
      await generate('recebi 2 cadeiras gamer do fornecedor');
      await submit();

      expect(service.create).toHaveBeenCalledWith(
        expect.objectContaining({ type: 'IN' })
      );
    });

    it('lists what it could not resolve', async () => {
      service.draft.mockReturnValue(
        of({
          ...DRAFT,
          unresolved: [
            { text: 'monitor LG', quantity: 2, reason: 'NOT_FOUND' as const, candidates: [] }
          ]
        })
      );

      await generate('vender 2 monitores LG');

      expect(issues()).toEqual([
        '"monitor LG": não encontrei esse produto no catálogo.'
      ]);
    });

    it('names the candidates when more than one product matches', async () => {
      service.draft.mockReturnValue(
        of({
          ...DRAFT,
          unresolved: [
            {
              text: 'cadeira',
              quantity: 1,
              reason: 'AMBIGUOUS' as const,
              candidates: [
                { inventoryId: 3, code: 'PRD-0003', name: 'Cadeira Gamer' },
                { inventoryId: 4, code: 'PRD-0004', name: 'Cadeira Comum' }
              ]
            }
          ]
        })
      );

      await generate('vender uma cadeira');

      expect(issues()).toEqual([
        '"cadeira": mais de um produto combina — Cadeira Gamer, Cadeira Comum. Escolha na lista de itens.'
      ]);
    });

    it('shows the server reason when the feature is not configured', async () => {
      rejectDraftWith(
        { message: 'O preenchimento por IA não está configurado neste ambiente.' },
        503
      );

      await generate('vender 2 cadeiras gamer');

      expect(promptFailure()).toBe(
        'O preenchimento por IA não está configurado neste ambiente.'
      );
    });

    it('shows a general warning when the extraction fails', async () => {
      rejectDraftWith({ message: 'qualquer coisa' }, 502);

      await generate('vender 2 cadeiras gamer');

      expect(promptFailure()).toBe(
        'Não foi possível interpretar o pedido. Tente novamente.'
      );
    });

    it('keeps the form usable after a failed draft', async () => {
      rejectDraftWith(null, 502);

      await generate('vender 2 cadeiras gamer');
      await fillValidForm();
      await submit();

      expect(service.create).toHaveBeenCalledTimes(1);
      expect(navigate).toHaveBeenCalledWith(['/billing/invoices']);
    });

    it('drops the warning of the previous attempt', async () => {
      rejectDraftWith(null, 502);
      await generate('vender 2 cadeiras gamer');
      expect(promptFailure()).not.toBe('');

      service.draft.mockReturnValue(new Subject<InvoiceDraft>());
      promptButton().click();
      await fixture.whenStable();

      expect(promptFailure()).toBe('');
    });

    it('blocks a second draft while the first is still running', async () => {
      service.draft.mockReturnValue(new Subject<InvoiceDraft>());

      await generate('vender 2 cadeiras gamer');
      promptButton().click();
      await fixture.whenStable();

      expect(service.draft).toHaveBeenCalledTimes(1);
      expect(text(promptButton())).toBe('Montando…');
    });
  });

  describe('next document', () => {
    const changeSeries = async (value: string) => {
      await fill('series', value);
      await fixture.whenStable();
    };

    it('asks the API for the next document when the screen opens', () => {
      expect(service.nextDocument).toHaveBeenCalledTimes(1);
      expect(service.nextDocument).toHaveBeenCalledWith(undefined);
    });

    it('opens with the series and number already filled', () => {
      expect(field<HTMLInputElement>('series').value).toBe('1');
      expect(field<HTMLInputElement>('number').value).toBe('7');
    });

    it('saves the suggestion without the user typing anything', async () => {
      await submit();

      expect(service.create).toHaveBeenCalledWith({
        series: 1,
        number: 7,
        type: 'OUT',
        items: []
      });
    });

    it('asks again for the number when the series changes', async () => {
      service.nextDocument.mockReturnValue(of({ series: 4, number: 1 }));

      await changeSeries('4');

      expect(service.nextDocument).toHaveBeenLastCalledWith(4);
      expect(field<HTMLInputElement>('number').value).toBe('1');
    });

    it('keeps the series the user typed', async () => {
      service.nextDocument.mockReturnValue(of({ series: 4, number: 1 }));

      await changeSeries('4');

      expect(field<HTMLInputElement>('series').value).toBe('4');
    });

    it('never overwrites a number the user typed', async () => {
      await fill('number', '99');

      service.nextDocument.mockReturnValue(of({ series: 4, number: 1 }));
      await changeSeries('4');

      expect(field<HTMLInputElement>('number').value).toBe('99');
    });

    it('asks nothing while the series is empty', async () => {
      await changeSeries('');

      expect(service.nextDocument).toHaveBeenCalledTimes(1);
    });

    it('leaves the fields empty when the suggestion fails', async () => {
      await mount(undefined, () => throwError(() => new HttpErrorResponse({ status: 500 })));

      expect(field<HTMLInputElement>('series').value).toBe('');
      expect(field<HTMLInputElement>('number').value).toBe('');
    });

    it('leaves the number empty when the series is exhausted', async () => {
      service.nextDocument.mockReturnValue(of({ series: 1, number: null }));

      await changeSeries('1');

      expect(field<HTMLInputElement>('number').value).toBe('');
    });

    it('still lets the user save after a failed suggestion', async () => {
      await mount(undefined, () => throwError(() => new HttpErrorResponse({ status: 500 })));

      await fillValidForm();
      await submit();

      expect(service.create).toHaveBeenCalledTimes(1);
      expect(navigate).toHaveBeenCalledWith(['/billing/invoices']);
    });
  });
});
