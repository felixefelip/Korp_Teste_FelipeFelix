import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router, convertToParamMap, provideRouter } from '@angular/router';
import { Observable, Subject, of, throwError } from 'rxjs';

import { FlashService } from '../../../shared/flash/flash.service';
import { Product } from '../product.model';
import { ProductService } from '../product.service';
import { ProductEdit } from './product-edit';

const EXISTING: Product = {
  id: 7,
  code: 'PRD-0007',
  name: 'Cadeira de escritório',
  unit: 'CX',
  price: 750.5,
  stock: 8
};

describe('ProductEdit', () => {
  let fixture: ComponentFixture<ProductEdit>;
  let service: {
    create: ReturnType<typeof vi.fn>;
    get: ReturnType<typeof vi.fn>;
    update: ReturnType<typeof vi.fn>;
  };
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

  const mount = async (
    load: () => Observable<Product> = () => of(EXISTING),
    id = '7'
  ) => {
    TestBed.resetTestingModule();

    service = {
      create: vi.fn(),
      get: vi.fn(load),
      update: vi.fn((productId: number, data: Omit<Product, 'id'>) =>
        of({ ...data, id: productId })
      )
    };

    flash = { error: vi.fn(), success: vi.fn() };

    await TestBed.configureTestingModule({
      imports: [ProductEdit],
      providers: [
        provideRouter([]),
        { provide: ProductService, useValue: service },
        { provide: FlashService, useValue: flash },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: convertToParamMap({ id }) } }
        }
      ]
    }).compileComponents();

    navigate = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);

    fixture = TestBed.createComponent(ProductEdit);
    await fixture.whenStable();
  };

  beforeEach(async () => {
    await mount();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('loading the product', () => {
    it('asks the API for the product in the route', () => {
      expect(service.get).toHaveBeenCalledWith(7);
    });

    it('fills every field with what came back', () => {
      expect(field<HTMLInputElement>('code').value).toBe('PRD-0007');
      expect(field<HTMLInputElement>('name').value).toBe('Cadeira de escritório');
      expect(field<HTMLSelectElement>('unit').value).toBe('CX');
      expect(field<HTMLInputElement>('price').value).toBe('750.5');
      expect(field<HTMLInputElement>('stock').value).toBe('8');
    });

    it('announces that it is editing, not creating', () => {
      expect(text(element().querySelector('.page__title'))).toBe('Editar produto');
      expect(text(submitButton())).toBe('Salvar alterações');
    });

    it('shows no error over a product that is still untouched', () => {
      expect(errors()).toEqual([]);
      expect(failure()).toBe('');
      expect(flash.error).not.toHaveBeenCalled();
    });

    it('waits for the product before showing the fields', async () => {
      await mount(() => new Subject<Product>());

      expect(text(element().querySelector('.form__status'))).toBe(
        'Carregando produto…'
      );
      expect(element().querySelector('#code')).toBeNull();
      expect(submitButton().disabled).toBe(true);
    });
  });

  describe('saving the changes', () => {
    it('sends the edited data to the update of that id', async () => {
      await fill('name', 'Cadeira gamer');
      await fill('price', '899.9');
      await submit();

      expect(service.update).toHaveBeenCalledWith(7, {
        code: 'PRD-0007',
        name: 'Cadeira gamer',
        unit: 'CX',
        price: 899.9,
        stock: 8
      });
    });

    it('never creates a second product', async () => {
      await submit();

      expect(service.create).not.toHaveBeenCalled();
      expect(service.update).toHaveBeenCalledTimes(1);
    });

    it('goes back to the listing after saving', async () => {
      await submit();

      expect(navigate).toHaveBeenCalledWith(['/inventory/products']);
    });

    it('still blocks the submit when a field was emptied', async () => {
      await fill('name', '');
      await submit();

      expect(service.update).not.toHaveBeenCalled();
      expect(errorOf('name')).toBe('Campo obrigatório.');
    });

    it('shows on the field the error the server pointed at', async () => {
      service.update.mockReturnValue(
        throwError(
          () =>
            new HttpErrorResponse({
              status: 400,
              error: { errors: { code: 'Campo obrigatório.' } }
            })
        )
      );

      await submit();

      expect(errorOf('code')).toBe('Campo obrigatório.');
      expect(navigate).not.toHaveBeenCalled();
    });

    it('warns without leaking the message of a 500', async () => {
      service.update.mockReturnValue(
        throwError(
          () =>
            new HttpErrorResponse({
              status: 500,
              error: { message: 'erro ao atualizar o produto' }
            })
        )
      );

      await submit();

      expect(failure()).toBe('Não foi possível salvar o produto. Tente novamente.');
      expect(navigate).not.toHaveBeenCalled();
    });
  });

  describe('product that cannot be loaded', () => {
    const failLoadWith = (status: number) =>
      mount(() => throwError(() => new HttpErrorResponse({ status })));

    it('goes back to the listing when the product does not exist', async () => {
      await failLoadWith(404);

      expect(navigate).toHaveBeenCalledWith(['/inventory/products']);
    });

    it('goes back to the listing when the API is down', async () => {
      await failLoadWith(500);

      expect(navigate).toHaveBeenCalledWith(['/inventory/products']);
    });

    it('says on the way out that the product does not exist', async () => {
      await failLoadWith(404);

      expect(flash.error).toHaveBeenCalledWith('Produto não encontrado.');
    });

    it('blames the API when the failure was not a 404', async () => {
      await failLoadWith(500);

      expect(flash.error).toHaveBeenCalledWith(
        'Não foi possível carregar o produto. Tente novamente.'
      );
    });

    it('warns once per failed load', async () => {
      await failLoadWith(404);

      expect(flash.error).toHaveBeenCalledTimes(1);
    });

    it('never shows the empty form while it leaves the screen', async () => {
      await failLoadWith(404);

      expect(element().querySelector('#code')).toBeNull();
      expect(text(element().querySelector('.form__status'))).toBe(
        'Carregando produto…'
      );
    });

    it('saves nothing on the way out', async () => {
      await failLoadWith(404);

      await submit();

      expect(service.update).not.toHaveBeenCalled();
      expect(service.create).not.toHaveBeenCalled();
    });
  });
});
