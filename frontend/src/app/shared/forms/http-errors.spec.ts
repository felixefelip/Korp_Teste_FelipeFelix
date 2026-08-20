import { FormControl, FormGroup } from '@angular/forms';

import { applyMessageErrorsToForm, isClientError } from './http-errors';

describe('isClientError', () => {
  it('accepts the 4xx range', () => {
    expect(isClientError(400)).toBe(true);
    expect(isClientError(409)).toBe(true);
    expect(isClientError(422)).toBe(true);
    expect(isClientError(499)).toBe(true);
  });

  it('rejects server errors', () => {
    expect(isClientError(500)).toBe(false);
    expect(isClientError(502)).toBe(false);
  });

  it('rejects the status of a request that never reached the server', () => {
    expect(isClientError(0)).toBe(false);
  });
});

describe('applyMessageErrorsToForm', () => {
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
