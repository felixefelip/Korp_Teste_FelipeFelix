import { Component, computed, effect, inject, input, output, signal } from '@angular/core';
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
  InvoiceDocument,
  InvoiceDraft,
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
  readonly draft = input<InvoiceDraft | null>(null);
  readonly suggestion = input<InvoiceDocument | null>(null);
  readonly products = input<CatalogProduct[]>([]);
  readonly productsFailed = input(false);
  readonly loading = input(false);
  readonly saving = input(false);
  readonly failure = input<ApiFailure | null>(null);
  readonly typeEditable = input(true);
  readonly submitLabel = input('Salvar nota fiscal');

  readonly save = output<InvoicePayload>();
  readonly seriesChange = output<number>();

  protected readonly types = INVOICE_TYPES;
  protected readonly typeLabels = INVOICE_TYPE_LABELS;
  protected readonly submitted = signal(false);
  protected readonly banner = signal<string | null>(null);

  protected readonly shortages = computed(() => this.value()?.shortages ?? []);

  protected readonly stockWarning = computed(() => {
    const shortages = this.shortages();

    if (!shortages.length) {
      return null;
    }

    const total = shortages.reduce((sum, shortage) => sum + shortage.required, 0);
    const products = shortages.map((shortage) => shortage.code).join(', ');

    return shortages.length === 1
      ? `Não foi possível imprimir: ${products} tem ${shortages[0].available} em estoque e a nota pede ${total}.`
      : `Não foi possível imprimir: o estoque não cobre ${products}.`;
  });

  protected readonly form = this.fb.group({
    series: this.fb.control<number | null>(null, [
      Validators.required,
      Validators.min(1),
      Validators.max(999)
    ]),
    number: this.fb.control<number | null>(null, [
      Validators.required,
      Validators.min(1),
      Validators.max(999999)
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
      this.form.patchValue({ series: invoice.series, number: invoice.number, type: invoice.type });
    });

    effect(() => {
      const suggestion = this.suggestion();

      if (!suggestion) {
        return;
      }

      const { series, number } = this.form.controls;

      if (series.pristine) {
        series.setValue(suggestion.series);
      }

      if (number.pristine) {
        number.setValue(suggestion.number);
      }
    });

    effect(() => {
      const draft = this.draft();

      if (!draft) {
        return;
      }

      this.form.setControl('items', newItemArray(draft.items));

      if (this.typeEditable()) {
        this.form.patchValue({ type: draft.type });
      }
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

  protected announceSeries(): void {
    const series = this.form.controls.series.value;

    if (series && series > 0) {
      this.seriesChange.emit(series);
    }
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

    const { series, number, type, items } = this.form.getRawValue();

    this.save.emit({
      series: series!,
      number: number!,
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
