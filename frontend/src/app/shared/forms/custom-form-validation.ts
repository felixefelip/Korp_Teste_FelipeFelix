import { Signal } from '@angular/core';
import { AbstractControl, FormGroup } from '@angular/forms';

export type ValidationMessage = (error: any) => string;

const SERVER_ERROR_KEY = 'server';
const FALLBACK = 'Valor inválido.';

export class CustomFormValidation {
  static readonly messages: Record<string, ValidationMessage> = {
    required: () => 'Campo obrigatório.',
    minlength: (error) => `Informe pelo menos ${error.requiredLength} caracteres.`,
    maxlength: (error) => `Limite de ${error.requiredLength} caracteres excedido.`,
    min: (error) =>
      error.min === 0
        ? 'O valor não pode ser negativo.'
        : `O valor mínimo é ${error.min}.`,
    max: (error) => `O valor máximo é ${error.max}.`,
    email: () => 'Informe um e-mail válido.',
    integer: () => 'Informe um número inteiro.'
  };

  static fieldErrorMessage(
    control: AbstractControl | null,
    overrides: Record<string, ValidationMessage> = {}
  ): string | null {
    if (!control || control.valid) {
      return null;
    }

    const errors = control.errors ?? {};

    if (typeof errors[SERVER_ERROR_KEY] === 'string') {
      return errors[SERVER_ERROR_KEY];
    }

    const key = Object.keys(errors)[0];

    if (!key) {
      return null;
    }

    const message = overrides[key] ?? CustomFormValidation.messages[key];

    return message ? message(errors[key]) : FALLBACK;
  }

  static controlPath(field: string): string {
    return field.replace(/\[(\d+)\]/g, '.$1');
  }

  static fieldErrorFor(
    form: FormGroup,
    submitted: Signal<boolean>,
    overrides: Record<string, ValidationMessage> = {}
  ): (field: string) => string | null {
    return (field) => {
      const control = form.get(CustomFormValidation.controlPath(field));

      if (!control || !(control.touched || submitted())) {
        return null;
      }

      return CustomFormValidation.fieldErrorMessage(control, overrides);
    };
  }

  static applyMessageErrorsToForm(form: FormGroup, errors: unknown): boolean {
    if (!errors || typeof errors !== 'object') {
      return false;
    }

    let anyKnownField = false;

    for (const [field, message] of Object.entries(errors)) {
      const control = form.get(CustomFormValidation.controlPath(field));

      if (control && typeof message === 'string') {
        control.setErrors({ [SERVER_ERROR_KEY]: message });
        anyKnownField = true;
      }
    }

    return anyKnownField;
  }
}
