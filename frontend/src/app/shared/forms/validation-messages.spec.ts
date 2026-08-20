import { FormControl, Validators } from '@angular/forms';

import { fieldErrorMessage } from './validation-messages';

describe('fieldErrorMessage', () => {
  const invalidControl = (
    value: unknown,
    validators: Parameters<FormControl['addValidators']>[0]
  ) => {
    const control = new FormControl(value, validators as never);
    control.updateValueAndValidity();
    return control;
  };

  it('returns null for a valid control', () => {
    expect(fieldErrorMessage(invalidControl('ok', Validators.required))).toBeNull();
  });

  it('returns null when there is no control', () => {
    expect(fieldErrorMessage(null)).toBeNull();
  });

  it('describes a required field', () => {
    expect(fieldErrorMessage(invalidControl('', Validators.required))).toBe(
      'Campo obrigatório.'
    );
  });

  it('takes the minimum length from the control, not from a fixed text', () => {
    expect(fieldErrorMessage(invalidControl('ab', Validators.minLength(3)))).toBe(
      'Informe pelo menos 3 caracteres.'
    );
    expect(fieldErrorMessage(invalidControl('ab', Validators.minLength(9)))).toBe(
      'Informe pelo menos 9 caracteres.'
    );
  });

  it('takes the maximum length from the control', () => {
    expect(fieldErrorMessage(invalidControl('abcdef', Validators.maxLength(4)))).toBe(
      'Limite de 4 caracteres excedido.'
    );
  });

  it('phrases a minimum of zero as a non-negative rule', () => {
    expect(fieldErrorMessage(invalidControl(-1, Validators.min(0)))).toBe(
      'O valor não pode ser negativo.'
    );
  });

  it('states the minimum when it is not zero', () => {
    expect(fieldErrorMessage(invalidControl(1, Validators.min(5)))).toBe(
      'O valor mínimo é 5.'
    );
  });

  it('falls back for a rule it has no copy for', () => {
    const control = new FormControl('');
    control.setErrors({ somethingElse: true });

    expect(fieldErrorMessage(control)).toBe('Valor inválido.');
  });

  it('prefers an override over the default copy', () => {
    const control = invalidControl('', Validators.required);

    expect(fieldErrorMessage(control, { required: () => 'Preencha aqui.' })).toBe(
      'Preencha aqui.'
    );
  });

  it('shows the server message ahead of any validator', () => {
    const control = invalidControl('', Validators.required);
    control.setErrors({ required: true, server: 'O código já existe.' });

    expect(fieldErrorMessage(control)).toBe('O código já existe.');
  });
});
