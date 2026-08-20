import { HttpStatusCode } from '@angular/common/http';
import { FormGroup } from '@angular/forms';

export function isClientError(status: number): boolean {
  return (
    status >= HttpStatusCode.BadRequest && status < HttpStatusCode.InternalServerError
  );
}

export function applyMessageErrorsToForm(form: FormGroup, errors: unknown): boolean {
  if (!errors || typeof errors !== 'object') {
    return false;
  }

  let anyKnownField = false;

  for (const [field, message] of Object.entries(errors)) {
    const control = form.get(field);

    if (control && typeof message === 'string') {
      control.setErrors({ server: message });
      anyKnownField = true;
    }
  }

  return anyKnownField;
}
