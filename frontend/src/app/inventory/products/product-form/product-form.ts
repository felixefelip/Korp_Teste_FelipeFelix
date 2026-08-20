import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';

import {
  CustomFormValidation,
  ValidationMessage
} from '../../../shared/forms/custom-form-validation';
import { isClientError } from '../../../shared/forms/http-errors';
import { CustomValidators } from '../../../shared/forms/validators';
import { ProductService } from '../product.service';

export const UNITS = ['UN', 'CX', 'PC', 'KG', 'L', 'M'];

const MESSAGE_OVERRIDES: Record<string, ValidationMessage> = {
  pattern: () => 'Use apenas letras, números e hífen.'
};

const GENERIC_FAILURE = 'Não foi possível salvar o produto. Tente novamente.';

@Component({
  selector: 'app-product-form',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './product-form.html',
  styleUrl: './product-form.scss'
})
export class ProductForm {
  private readonly fb = inject(FormBuilder);
  private readonly productService = inject(ProductService);
  private readonly router = inject(Router);

  protected readonly units = UNITS;
  protected readonly submitted = signal(false);
  protected readonly saving = signal(false);
  protected readonly failure = signal<string | null>(null);

  protected readonly form = this.fb.group({
    code: this.fb.nonNullable.control(this.productService.nextCode(), [
      Validators.required,
      Validators.pattern(/^[A-Za-z0-9-]+$/)
    ]),
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
    this.submitted,
    MESSAGE_OVERRIDES
  );

  constructor() {
    this.productService.list().subscribe({
      next: () => {
        const control = this.form.controls.code;

        if (control.pristine) {
          control.setValue(this.productService.nextCode());
        }
      },
      error: () => {}
    });
  }

  protected save(): void {
    this.submitted.set(true);
    this.failure.set(null);

    if (this.form.invalid || this.saving()) {
      this.form.markAllAsTouched();
      return;
    }

    const { code, name, unit, price, stock } = this.form.getRawValue();

    this.saving.set(true);

    this.productService
      .create({ code, name: name.trim(), unit, price: price!, stock: stock! })
      .subscribe({
        next: () => {
          this.saving.set(false);
          this.router.navigate(['/inventory/products']);
        },
        error: (response: HttpErrorResponse) => {
          this.saving.set(false);
          this.handleServerFailure(response);
        }
      });
  }

  private handleServerFailure(response: HttpErrorResponse): void {
    if (!isClientError(response.status)) {
      this.failure.set(GENERIC_FAILURE);
      return;
    }

    const errors = response.error?.errors;

    if (CustomFormValidation.applyMessageErrorsToForm(this.form, errors)) {
      this.failure.set(null);
      return;
    }

    const message = response.error?.message;

    this.failure.set(typeof message === 'string' && message ? message : GENERIC_FAILURE);
  }
}
