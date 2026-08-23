import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';

import { CatalogProduct } from '../catalog.model';
import { CustomFormValidation } from '../../../shared/forms/custom-form-validation';
import { ApiFailure } from '../../../shared/forms/http-errors';
import { InvoiceItemsForm, newItemArray } from '../invoice-items-form/invoice-items-form';
import {
  INVOICE_TYPES,
  INVOICE_TYPE_LABELS,
  Invoice,
  InvoicePayload,
  InvoiceType
} from '../invoice.model';

export const SAVE_FAILURE = 'Não foi possível salvar a nota fiscal. Tente novamente.';

@Component({
  selector: 'app-invoice-form',
  imports: [InvoiceItemsForm, ReactiveFormsModule, RouterLink],
  templateUrl: './invoice-form.html',
  styleUrl: './invoice-form.scss'
})
export class InvoiceForm {
  private readonly fb = inject(FormBuilder);

  readonly value = input<Invoice | null>(null);
  readonly products = input<CatalogProduct[]>([]);
  readonly productsFailed = input(false);
  readonly loading = input(false);
  readonly saving = input(false);
  readonly failure = input<ApiFailure | null>(null);
  readonly typeEditable = input(true);
  readonly submitLabel = input('Salvar nota fiscal');

  readonly save = output<InvoicePayload>();

  protected readonly types = INVOICE_TYPES;
  protected readonly typeLabels = INVOICE_TYPE_LABELS;
  protected readonly submitted = signal(false);
  protected readonly banner = signal<string | null>(null);

  protected readonly form = this.fb.group({
    number: this.fb.nonNullable.control('', [
      Validators.required,
      Validators.maxLength(30)
    ]),
    type: this.fb.nonNullable.control<InvoiceType>(
      INVOICE_TYPES[0],
      Validators.required
    ),
    items: newItemArray()
  });

  protected readonly fieldError = CustomFormValidation.fieldErrorFor(
    this.form,
    this.submitted
  );

  constructor() {
    effect(() => {
      const invoice = this.value();

      if (!invoice) {
        return;
      }

      this.form.setControl('items', newItemArray(invoice.items ?? []));
      this.form.patchValue({ number: invoice.number, type: invoice.type });
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

  protected get items() {
    return this.form.controls.items;
  }

  protected currentTypeLabel(): string {
    const type = this.form.controls.type.value;

    return INVOICE_TYPE_LABELS[type] ?? type ?? '';
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

    const { number, type, items } = this.form.getRawValue();

    this.save.emit({
      number: number.trim(),
      ...(this.typeEditable() ? { type } : {}),
      items: items.map((item) => ({
        inventoryId: item.inventoryId!,
        code: item.code,
        name: item.name,
        unit: item.unit,
        quantity: item.quantity!,
        unitPrice: item.unitPrice!
      }))
    });
  }
}
