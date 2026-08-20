import { FormControl } from '@angular/forms';

import { integer } from './validators';

describe('integer', () => {
  const validate = (value: unknown) => integer(new FormControl(value));

  it('accepts an integer', () => {
    expect(validate(7)).toBeNull();
  });

  it('accepts zero', () => {
    expect(validate(0)).toBeNull();
  });

  it('accepts a negative integer', () => {
    expect(validate(-3)).toBeNull();
  });

  it('accepts an empty control, leaving that to required', () => {
    expect(validate(null)).toBeNull();
    expect(validate('')).toBeNull();
  });

  it('rejects a fractional number', () => {
    expect(validate(2.5)).toEqual({ notAnInteger: true });
  });

  it('rejects a value that is not a number', () => {
    expect(validate('muitos')).toEqual({ notAnInteger: true });
  });
});
