import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';

import { CustomFormValidation } from '../../../shared/forms/custom-form-validation';
import { ApiFailure } from '../../../shared/forms/http-errors';
import { CustomValidators } from '../../../shared/forms/validators';
import { Product } from '../product.model';

export const UNITS = ['UN', 'CX', 'PC', 'KG', 'L', 'M'];

export const SAVE_FAILURE = 'Não foi possível salvar o produto. Tente novamente.';

export type ProductPayload = Omit<Product, 'id'>;

@Component({
  selector: 'app-product-form',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './product-form.html',
  styleUrl: './product-form.scss'
})
export class ProductForm {
  private readonly fb = inject(FormBuilder);

  readonly value = input<Product | null>(null);
  readonly loading = input(false);
  readonly saving = input(false);
  readonly failure = input<ApiFailure | null>(null);
  readonly stockLabel = input('Estoque');
  readonly submitLabel = input('Salvar produto');

  readonly save = output<ProductPayload>();

  protected readonly units = UNITS;
  protected readonly submitted = signal(false);
  protected readonly banner = signal<string | null>(null);

  protected readonly form = this.fb.group({
    code: this.fb.nonNullable.control('', Validators.required),
    name: this.fb.nonNullable.control('', [
      Validators.required,
      Validators.minLength(3),
      Validators.maxLength(120)
    ]),
    unit: this.fb.nonNullable.control(UNITS[0], Validators.required),
    price: this.fb.control<number | null>(null, [
      Validators.required,
      Validators.min(0)
    ]),
    stock: this.fb.control<number | null>(0, [
      Validators.required,
      Validators.min(0),
      CustomValidators.integer
    ])
  });

  protected readonly fieldError = CustomFormValidation.fieldErrorFor(
    this.form,
    this.submitted
  );

  constructor() {
    effect(() => {
      const product = this.value();

      if (!product) {
        return;
      }

      this.form.setValue({
        code: product.code,
        name: product.name,
        unit: product.unit,
        price: product.price,
        stock: product.stock
      });
    });

    effect(() => {
      const failure = this.failure();

      if (!failure) {
        this.banner.set(null);
        return;
      }

      const applied = CustomFormValidation.applyMessageErrorsToForm(
        this.form,
        failure.fieldErrors
      );

      this.banner.set(applied ? null : failure.message);
    });
  }

  protected submit(): void {
    if (this.loading()) {
      return;
    }

    this.submitted.set(true);

    if (this.form.invalid || this.saving()) {
      this.form.markAllAsTouched();
      return;
    }

    const { code, name, unit, price, stock } = this.form.getRawValue();

    this.save.emit({ code, name: name.trim(), unit, price: price!, stock: stock! });
  }
}
