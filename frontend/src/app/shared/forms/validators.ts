import { AbstractControl, ValidationErrors } from '@angular/forms';

export const CustomValidators = {
  integer(control: AbstractControl): ValidationErrors | null {
    const value = control.value;
    return value === null || value === '' || Number.isInteger(value)
      ? null
      : { integer: true };
  }
};
