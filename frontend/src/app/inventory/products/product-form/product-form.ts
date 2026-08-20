import { HttpErrorResponse, HttpStatusCode } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';

import { FlashService } from '../../../shared/flash/flash.service';
import { CustomFormValidation } from '../../../shared/forms/custom-form-validation';
import { isClientError } from '../../../shared/forms/http-errors';
import { CustomValidators } from '../../../shared/forms/validators';
import { ProductService } from '../product.service';

export const UNITS = ['UN', 'CX', 'PC', 'KG', 'L', 'M'];

const GENERIC_FAILURE = 'Não foi possível salvar o produto. Tente novamente.';
const NOT_FOUND_FAILURE = 'Produto não encontrado.';
const LOAD_FAILURE = 'Não foi possível carregar o produto. Tente novamente.';

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
  private readonly route = inject(ActivatedRoute);
  private readonly flash = inject(FlashService);

  private readonly productId = Number(this.route.snapshot.paramMap.get('id')) || null;

  protected readonly editing = this.productId !== null;
  protected readonly units = UNITS;
  protected readonly submitted = signal(false);
  protected readonly saving = signal(false);
  protected readonly loading = signal(this.productId !== null);
  protected readonly failure = signal<string | null>(null);

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
    if (this.productId !== null) {
      this.load(this.productId);
    }
  }

  protected save(): void {
    if (this.loading()) {
      return;
    }

    this.submitted.set(true);
    this.failure.set(null);

    if (this.form.invalid || this.saving()) {
      this.form.markAllAsTouched();
      return;
    }

    const { code, name, unit, price, stock } = this.form.getRawValue();
    const product = { code, name: name.trim(), unit, price: price!, stock: stock! };

    this.saving.set(true);

    const saved =
      this.productId === null
        ? this.productService.create(product)
        : this.productService.update(this.productId, product);

    saved.subscribe({
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

  private load(id: number): void {
    this.productService.get(id).subscribe({
      next: (product) => {
        this.form.setValue({
          code: product.code,
          name: product.name,
          unit: product.unit,
          price: product.price,
          stock: product.stock
        });
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.flash.error(
          response.status === HttpStatusCode.NotFound ? NOT_FOUND_FAILURE : LOAD_FAILURE
        );
        this.router.navigate(['/inventory/products']);
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
