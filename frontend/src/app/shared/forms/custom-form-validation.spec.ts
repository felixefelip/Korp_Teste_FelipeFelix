import { signal } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';

import { CustomFormValidation } from './custom-form-validation';

const { applyMessageErrorsToForm, fieldErrorFor, fieldErrorMessage } =
  CustomFormValidation;

describe('CustomFormValidation.fieldErrorMessage', () => {
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

describe('CustomFormValidation.fieldErrorFor', () => {
  const newForm = () =>
    new FormGroup({
      name: new FormControl('', Validators.required)
    });

  it('stays quiet while the field is untouched and nothing was submitted', () => {
    const fieldError = fieldErrorFor(newForm(), signal(false));

    expect(fieldError('name')).toBeNull();
  });

  it('speaks once the field is touched', () => {
    const form = newForm();
    const fieldError = fieldErrorFor(form, signal(false));

    form.controls.name.markAsTouched();

    expect(fieldError('name')).toBe('Campo obrigatório.');
  });

  it('speaks for every field once the form was submitted', () => {
    const fieldError = fieldErrorFor(newForm(), signal(true));

    expect(fieldError('name')).toBe('Campo obrigatório.');
  });

  it('follows the signal as it changes', () => {
    const submitted = signal(false);
    const fieldError = fieldErrorFor(newForm(), submitted);

    expect(fieldError('name')).toBeNull();

    submitted.set(true);

    expect(fieldError('name')).toBe('Campo obrigatório.');
  });

  it('returns null for a field the form does not have', () => {
    const fieldError = fieldErrorFor(newForm(), signal(true));

    expect(fieldError('nonexistent')).toBeNull();
  });

  it('passes the overrides through', () => {
    const fieldError = fieldErrorFor(newForm(), signal(true), {
      required: () => 'Preencha aqui.'
    });

    expect(fieldError('name')).toBe('Preencha aqui.');
  });
});

describe('CustomFormValidation.applyMessageErrorsToForm', () => {
  const newForm = () =>
    new FormGroup({
      code: new FormControl(''),
      price: new FormControl(0)
    });

  it('marks the control the server pointed at', () => {
    const form = newForm();

    expect(applyMessageErrorsToForm(form, { code: 'Campo obrigatório.' })).toBe(true);
    expect(form.controls.code.errors).toEqual({ server: 'Campo obrigatório.' });
  });

  it('marks every known field at once', () => {
    const form = newForm();

    applyMessageErrorsToForm(form, { code: 'Campo obrigatório.', price: 'Valor inválido.' });

    expect(form.controls.code.errors).toEqual({ server: 'Campo obrigatório.' });
    expect(form.controls.price.errors).toEqual({ server: 'Valor inválido.' });
  });

  it('reports false when no field matches a control', () => {
    const form = newForm();

    expect(applyMessageErrorsToForm(form, { unknownField: 'Campo obrigatório.' })).toBe(false);
    expect(form.valid).toBe(true);
  });

  it('ignores a message that is not a string', () => {
    const form = newForm();

    expect(applyMessageErrorsToForm(form, { code: { nested: true } })).toBe(false);
    expect(form.controls.code.errors).toBeNull();
  });

  it('reports false when the body is not an object', () => {
    expect(applyMessageErrorsToForm(newForm(), undefined)).toBe(false);
    expect(applyMessageErrorsToForm(newForm(), null)).toBe(false);
    expect(applyMessageErrorsToForm(newForm(), '<html>502</html>')).toBe(false);
  });
});
