import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router, convertToParamMap, provideRouter } from '@angular/router';
import { Observable, Subject, of, throwError } from 'rxjs';

import { Movement, MovementPayload } from '../movement.model';
import { MovementService } from '../movement.service';
import { MovementNew } from './movement-new';

const CREATED: Movement = {
  id: 4,
  productId: 7,
  type: 'in',
  origin: 'adjustment',
  quantity: 10,
  confirmed: true,
  invoiceItemId: null
};

describe('MovementNew', () => {
  let fixture: ComponentFixture<MovementNew>;
  let service: { create: ReturnType<typeof vi.fn> };
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

  const errorOf = (id: string) =>
    text(field(id).parentElement?.querySelector('.field-error'));

  const mount = async (
    create: () => Observable<Movement> = () => of(CREATED),
    id = '7'
  ) => {
    TestBed.resetTestingModule();

    service = { create: vi.fn(create) };

    await TestBed.configureTestingModule({
      imports: [MovementNew],
      providers: [
        provideRouter([]),
        { provide: MovementService, useValue: service },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: convertToParamMap({ id }) } }
        }
      ]
    }).compileComponents();

    navigate = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);

    fixture = TestBed.createComponent(MovementNew);
    await fixture.whenStable();
  };

  const fillValidForm = async () => {
    await fill('type', 'in');
    await fill('quantity', '10');
  };

  beforeEach(async () => {
    await mount();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  it('announces that it is creating', () => {
    expect(text(element().querySelector('.page__title'))).toBe('Nova movimentação');
    expect(text(element().querySelector('button[type="submit"]'))).toBe(
      'Salvar movimentação'
    );
  });

  it('opens with an empty form, never loading', () => {
    expect(field<HTMLInputElement>('quantity').value).toBe('');
    expect(element().querySelector('.form__status')).toBeNull();
  });

  describe('saving', () => {
    it('sends the movement to the product in the route', async () => {
      await fillValidForm();
      await submit();

      expect(service.create).toHaveBeenCalledWith(7, {
        type: 'in',
        quantity: 10,
        confirmed: false
      });
    });

    it('goes back to the ledger of that product', async () => {
      await fillValidForm();
      await submit();

      expect(navigate).toHaveBeenCalledWith(['/inventory/products', 7, 'movements']);
    });

    it('does not call the API for an invalid form', async () => {
      await submit();

      expect(service.create).not.toHaveBeenCalled();
      expect(navigate).not.toHaveBeenCalled();
    });

    it('says it is saving while the API does not answer', async () => {
      await mount(() => new Subject<Movement>());
      await fillValidForm();
      await submit();

      expect(text(element().querySelector('button[type="submit"]'))).toBe('Salvando…');
    });
  });

  describe('when the API refuses', () => {
    it('shows the field error it named', async () => {
      await mount(() =>
        throwError(
          () =>
            new HttpErrorResponse({
              status: 400,
              error: { errors: { quantity: 'Quantidade acima do saldo.' } }
            })
        )
      );
      await fillValidForm();
      await submit();

      expect(errorOf('quantity')).toBe('Quantidade acima do saldo.');
    });

    it('shows a banner when it names no field', async () => {
      await mount(() => throwError(() => new HttpErrorResponse({ status: 500 })));
      await fillValidForm();
      await submit();

      expect(failure()).toBe(
        'Não foi possível salvar a movimentação. Tente novamente.'
      );
      expect(navigate).not.toHaveBeenCalled();
    });

    it('keeps what was typed so it can be fixed', async () => {
      await mount(() => throwError(() => new HttpErrorResponse({ status: 500 })));
      await fillValidForm();
      await submit();

      expect(field<HTMLInputElement>('quantity').value).toBe('10');
    });
  });
});
