import { AbstractControl, ValidationErrors } from '@angular/forms';

export function integer(control: AbstractControl): ValidationErrors | null {
  const value = control.value;
  return value === null || value === '' || Number.isInteger(value)
    ? null
    : { notAnInteger: true };
}
