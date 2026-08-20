import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { Invoice } from '../invoice.model';
import { InvoiceForm, InvoicePayload } from './invoice-form';

const EXISTING: Invoice = { id: 7, number: 'NF-0007', status: 'CLOSED' };

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

  const fillValidForm = async () => {
    await fill('number', 'NF-0006');
    await fill('status', 'CLOSED');
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [InvoiceForm],
      providers: [provideRouter([])]
    }).compileComponents();

    fixture = TestBed.createComponent(InvoiceForm);
    saved = [];
    fixture.componentInstance.save.subscribe((invoice) => saved.push(invoice));
    await fixture.whenStable();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('initial state', () => {
    it('starts with an empty number', () => {
      expect(field<HTMLInputElement>('number').value).toBe('');
    });

    it('starts with the invoice open', () => {
      expect(field<HTMLSelectElement>('status').value).toBe('OPEN');
    });

    it('lists the available statuses in Portuguese', () => {
      const options = Array.from(field<HTMLSelectElement>('status').options).map(
        (option) => [option.value, text(option)]
      );

      expect(options).toEqual([
        ['OPEN', 'Aberta'],
        ['CLOSED', 'Fechada']
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
  });

  describe('validation', () => {
    it('blocks the submit and reveals the required field errors', async () => {
      await submit();

      expect(saved).toEqual([]);
      expect(errorOf('number')).toBe('Campo obrigatório.');
      expect(errorOf('status')).toBe('');
    });

    it('demands the number once the field is emptied', async () => {
      await fill('number', 'NF-0006');
      await fill('number', '');

      expect(errorOf('number')).toBe('Campo obrigatório.');
    });

    it('rejects a number longer than the 30 characters the API accepts', async () => {
      await fill('number', 'N'.repeat(31));

      expect(errorOf('number')).toBe('Limite de 30 caracteres excedido.');
    });

    it('accepts a number with exactly 30 characters', async () => {
      await fill('number', 'N'.repeat(30));
      await submit();

      expect(errorOf('number')).toBe('');
      expect(saved.length).toBe(1);
    });

    it('accepts a number already used by another invoice', async () => {
      await fillValidForm();
      await fill('number', 'NF-0001');
      await submit();

      expect(errorOf('number')).toBe('');
      expect(saved).toEqual([expect.objectContaining({ number: 'NF-0001' })]);
    });

    it('marks the invalid field visually', async () => {
      await submit();
      expect(field('number').classList).toContain('field--error');
    });

    it('clears the error as soon as the field is fixed', async () => {
      await submit();
      expect(errorOf('number')).toBe('Campo obrigatório.');

      await fill('number', 'NF-0006');
      expect(errorOf('number')).toBe('');
    });
  });

  describe('emitting what was filled', () => {
    it('hands over the filled data', async () => {
      await fillValidForm();
      await submit();

      expect(saved).toEqual([{ number: 'NF-0006', status: 'CLOSED' }]);
    });

    it('trims surrounding spaces from the number', async () => {
      await fill('number', '   NF-0007   ');
      await submit();

      expect(saved).toEqual([expect.objectContaining({ number: 'NF-0007' })]);
    });

    it('hands over the open status when it is untouched', async () => {
      await fill('number', 'NF-0006');
      await submit();

      expect(saved).toEqual([expect.objectContaining({ status: 'OPEN' })]);
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

  describe('the invoice being edited', () => {
    it('fills every field with the value it was given', async () => {
      await setInput('value', EXISTING);

      expect(field<HTMLInputElement>('number').value).toBe('NF-0007');
      expect(field<HTMLSelectElement>('status').value).toBe('CLOSED');
    });

    it('shows no error over an invoice that is still untouched', async () => {
      await setInput('value', EXISTING);

      expect(errors()).toEqual([]);
      expect(banner()).toBe('');
    });

    it('hands over the edited data', async () => {
      await setInput('value', EXISTING);
      await fill('number', 'NF-0042');
      await fill('status', 'OPEN');
      await submit();

      expect(saved).toEqual([{ number: 'NF-0042', status: 'OPEN' }]);
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
      await rejectWith({ number: 'Limite de 30 caracteres excedido.' });

      expect(errorOf('number')).toBe('Limite de 30 caracteres excedido.');
    });

    it('shows a phrase the frontend has no copy for', async () => {
      await fillValidForm();
      await rejectWith({ status: 'Esta nota já foi transmitida.' });

      expect(errorOf('status')).toBe('Esta nota já foi transmitida.');
    });

    it('drops the server error as soon as the field is fixed', async () => {
      await fillValidForm();
      await rejectWith({ number: 'Campo obrigatório.' });
      expect(errorOf('number')).toBe('Campo obrigatório.');

      await fill('number', 'NF-0042');
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
  });
});
