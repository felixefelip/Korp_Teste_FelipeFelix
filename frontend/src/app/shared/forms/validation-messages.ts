import { AbstractControl } from '@angular/forms';

export type ValidationMessage = (error: any) => string;

export const VALIDATION_MESSAGES: Record<string, ValidationMessage> = {
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

const FALLBACK = 'Valor inválido.';

export function fieldErrorMessage(
  control: AbstractControl | null,
  overrides: Record<string, ValidationMessage> = {}
): string | null {
  if (!control || control.valid) {
    return null;
  }

  const errors = control.errors ?? {};

  if (typeof errors['server'] === 'string') {
    return errors['server'];
  }

  const key = Object.keys(errors)[0];

  if (!key) {
    return null;
  }

  const message = overrides[key] ?? VALIDATION_MESSAGES[key];

  return message ? message(errors[key]) : FALLBACK;
}
