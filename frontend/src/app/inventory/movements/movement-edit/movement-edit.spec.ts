import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router, convertToParamMap, provideRouter } from '@angular/router';
import { Observable, Subject, of, throwError } from 'rxjs';

import { FlashService } from '../../../shared/flash/flash.service';
import { Movement } from '../movement.model';
import { MovementService } from '../movement.service';
import { MovementEdit } from './movement-edit';

const EXISTING: Movement = {
  id: 4,
  productId: 7,
  type: 'out',
  origin: 'adjustment',
  quantity: 3,
  confirmed: true,
  invoiceItemId: null
};

const FROM_INVOICE: Movement = {
  ...EXISTING,
  origin: 'sale',
  invoiceItemId: 33
};

const BACK_LINK = ['/inventory/products', 7, 'movements'];

describe('MovementEdit', () => {
  let fixture: ComponentFixture<MovementEdit>;
  let service: { get: ReturnType<typeof vi.fn>; update: ReturnType<typeof vi.fn> };
  let flash: { error: ReturnType<typeof vi.fn>; success: ReturnType<typeof vi.fn> };
  let navigate: ReturnType<typeof vi.spyOn>;

  const element = () => fixture.nativeElement as HTMLElement;

  const field = <T extends HTMLElement>(id: string) =>
    element().querySelector<T>(`#${id}`)!;

  const text = (el: Element | null | undefined) =>
    (el?.textContent ?? '').replace(/[  ]/g, ' ').trim();

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

  const failure = () => text(element().querySelector('.form__failure'));

  const mount = async (
    load: () => Observable<Movement> = () => of(EXISTING),
    update: () => Observable<Movement> = () => of(EXISTING)
  ) => {
    TestBed.resetTestingModule();

    service = { get: vi.fn(load), update: vi.fn(update) };
    flash = { error: vi.fn(), success: vi.fn() };

    await TestBed.configureTestingModule({
      imports: [MovementEdit],
      providers: [
        provideRouter([]),
        { provide: MovementService, useValue: service },
        { provide: FlashService, useValue: flash },
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: { paramMap: convertToParamMap({ id: '7', movementId: '4' }) }
          }
        }
      ]
    }).compileComponents();

    navigate = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);

    fixture = TestBed.createComponent(MovementEdit);
    await fixture.whenStable();
  };

  beforeEach(async () => {
    await mount();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('loading the movement', () => {
    it('asks the API for the movement of that product', () => {
      expect(service.get).toHaveBeenCalledWith(7, 4);
    });

    it('fills every field with what came back', () => {
      expect(field<HTMLSelectElement>('type').value).toBe('out');
      expect(field<HTMLInputElement>('quantity').value).toBe('3');
      expect(field<HTMLInputElement>('confirmed').checked).toBe(true);
    });

    it('announces that it is editing, not creating', () => {
      expect(text(element().querySelector('.page__title'))).toBe(
        'Editar movimentação'
      );
      expect(text(element().querySelector('button[type="submit"]'))).toBe(
        'Salvar alterações'
      );
    });

    it('waits for the movement before showing the fields', async () => {
      await mount(() => new Subject<Movement>());

      expect(element().querySelector('#quantity')).toBeNull();
      expect(text(element().querySelector('.form__status'))).toBe(
        'Carregando movimentação…'
      );
    });
  });

  describe('a movement born from an invoice', () => {
    it('never opens for editing', async () => {
      await mount(() => of(FROM_INVOICE));

      expect(flash.error).toHaveBeenCalledWith(
        'Movimentações geradas por notas fiscais não podem ser alteradas.'
      );
      expect(navigate).toHaveBeenCalledWith(BACK_LINK);
    });
  });

  describe('saving the changes', () => {
    it('sends the edited data to that movement', async () => {
      await fill('quantity', '9');
      await submit();

      expect(service.update).toHaveBeenCalledWith(7, 4, {
        type: 'out',
        quantity: 9,
        confirmed: true
      });
    });

    it('goes back to the ledger of that product', async () => {
      await submit();

      expect(navigate).toHaveBeenCalledWith(BACK_LINK);
    });
  });

  describe('when the API fails', () => {
    it('flashes and leaves when the movement is gone', async () => {
      await mount(() => throwError(() => new HttpErrorResponse({ status: 404 })));

      expect(flash.error).toHaveBeenCalledWith('Movimentação não encontrada.');
      expect(navigate).toHaveBeenCalledWith(BACK_LINK);
    });

    it('flashes and leaves when the load breaks', async () => {
      await mount(() => throwError(() => new HttpErrorResponse({ status: 500 })));

      expect(flash.error).toHaveBeenCalledWith(
        'Não foi possível carregar a movimentação. Tente novamente.'
      );
    });

    it('shows the conflict of an invoice-born movement on the banner', async () => {
      await mount(
        () => of(EXISTING),
        () =>
          throwError(
            () =>
              new HttpErrorResponse({
                status: 409,
                error: {
                  message:
                    'Movimentações geradas por notas fiscais não podem ser alteradas.'
                }
              })
          )
      );
      await submit();

      expect(failure()).toBe(
        'Movimentações geradas por notas fiscais não podem ser alteradas.'
      );
      expect(navigate).not.toHaveBeenCalled();
    });
  });
});
