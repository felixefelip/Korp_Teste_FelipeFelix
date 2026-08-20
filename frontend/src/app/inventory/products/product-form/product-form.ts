import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import {
  AbstractControl,
  FormBuilder,
  ReactiveFormsModule,
  ValidationErrors,
  Validators
} from '@angular/forms';
import { Router, RouterLink } from '@angular/router';

import { ServerErrors } from '../product.model';
import { ProductService } from '../product.service';

export const UNITS = ['UN', 'CX', 'PC', 'KG', 'L', 'M'];

function integer(control: AbstractControl): ValidationErrors | null {
  const value = control.value;
  return value === null || value === '' || Number.isInteger(value)
    ? null
    : { notAnInteger: true };
}

const MESSAGES: Record<string, string> = {
  required: 'Campo obrigatório.',
  minlength: 'Informe pelo menos 3 caracteres.',
  maxlength: 'Limite de 120 caracteres excedido.',
  pattern: 'Use apenas letras, números e hífen.',
  min: 'O valor não pode ser negativo.',
  notAnInteger: 'Informe um número inteiro.'
};

const SERVER_MESSAGES: Record<string, string> = {
  obrigatorio: 'Campo obrigatório.',
  'tipo invalido': 'Valor inválido.',
  'nao pode ser menor que 0': 'O valor não pode ser negativo.'
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
      integer
    ])
  });

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

  protected error(field: string): string | null {
    const control = this.form.get(field);

    if (!control || control.valid || !(control.touched || this.submitted())) {
      return null;
    }

    const errors = control.errors ?? {};

    if (typeof errors['server'] === 'string') {
      return errors['server'];
    }

    return MESSAGES[Object.keys(errors)[0]] ?? 'Valor inválido.';
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
    const errors = response.error?.errors as ServerErrors | undefined;

    if (!errors) {
      this.failure.set(response.error?.message ?? GENERIC_FAILURE);
      return;
    }

    let anyKnownField = false;

    for (const [field, message] of Object.entries(errors)) {
      const control = this.form.get(field);

      if (control) {
        control.setErrors({ server: this.translate(message) });
        anyKnownField = true;
      }
    }

    this.failure.set(anyKnownField ? null : GENERIC_FAILURE);
  }

  private translate(message: string): string {
    return (
      SERVER_MESSAGES[message] ??
      `${message.charAt(0).toUpperCase()}${message.slice(1)}.`
    );
  }
}
