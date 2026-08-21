import { CurrencyPipe } from '@angular/common';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import {
  FormBuilder,
  FormControl,
  FormGroup,
  ReactiveFormsModule,
  Validators
} from '@angular/forms';
import { RouterLink } from '@angular/router';

import { Product } from '../../../inventory/products/product.model';
import { CustomFormValidation } from '../../../shared/forms/custom-form-validation';
import { ApiFailure } from '../../../shared/forms/http-errors';
import { CustomValidators } from '../../../shared/forms/validators';
import {
  INVOICE_STATUSES,
  INVOICE_STATUS_LABELS,
  Invoice,
  InvoiceItem,
  InvoicePayload,
  InvoiceStatus
} from '../invoice.model';

export const SAVE_FAILURE = 'Não foi possível salvar a nota fiscal. Tente novamente.';

type ItemGroup = FormGroup<{
  inventoryId: FormControl<number | null>;
  code: FormControl<string>;
  name: FormControl<string>;
  unit: FormControl<string>;
  quantity: FormControl<number | null>;
  unitPrice: FormControl<number | null>;
}>;

@Component({
  selector: 'app-invoice-form',
  imports: [CurrencyPipe, ReactiveFormsModule, RouterLink],
  templateUrl: './invoice-form.html',
  styleUrl: './invoice-form.scss'
})
export class InvoiceForm {
  private readonly fb = inject(FormBuilder);

  readonly value = input<Invoice | null>(null);
  readonly products = input<Product[]>([]);
  readonly productsFailed = input(false);
  readonly loading = input(false);
  readonly saving = input(false);
  readonly failure = input<ApiFailure | null>(null);
  readonly submitLabel = input('Salvar nota fiscal');

  readonly save = output<InvoicePayload>();

  protected readonly statuses = INVOICE_STATUSES;
  protected readonly submitted = signal(false);
  protected readonly banner = signal<string | null>(null);

  protected readonly form = this.fb.group({
    number: this.fb.nonNullable.control('', [
      Validators.required,
      Validators.maxLength(30)
    ]),
    status: this.fb.nonNullable.control<InvoiceStatus>(
      INVOICE_STATUSES[0],
      Validators.required
    ),
    items: this.fb.array<ItemGroup>([])
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

      this.items.clear();

      for (const item of invoice.items ?? []) {
        this.items.push(this.itemGroup(item));
      }

      this.form.patchValue({ number: invoice.number, status: invoice.status });
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

  protected statusLabel(status: InvoiceStatus): string {
    return INVOICE_STATUS_LABELS[status] ?? status;
  }

  protected itemError(index: number, field: string): string | null {
    return this.fieldError(`items.${index}.${field}`);
  }

  protected addItem(): void {
    this.items.push(this.itemGroup());
  }

  protected removeItem(index: number): void {
    this.items.removeAt(index);
  }

  protected fillFromProduct(index: number): void {
    const group = this.items.at(index);
    const product = this.products().find(
      (candidate) => candidate.id === group.controls.inventoryId.value
    );

    if (!product) {
      return;
    }

    group.patchValue({
      code: product.code,
      name: product.name,
      unit: product.unit,
      unitPrice: product.price
    });
  }

  protected itemTotal(index: number): number {
    const { quantity, unitPrice } = this.items.at(index).getRawValue();

    return (quantity ?? 0) * (unitPrice ?? 0);
  }

  protected invoiceTotal(): number {
    return this.items.controls.reduce(
      (total, _item, index) => total + this.itemTotal(index),
      0
    );
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

    const { number, status, items } = this.form.getRawValue();

    this.save.emit({
      number: number.trim(),
      status,
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

  private itemGroup(item?: InvoiceItem): ItemGroup {
    return this.fb.group({
      inventoryId: this.fb.control<number | null>(
        item?.inventoryId ?? null,
        Validators.required
      ),
      code: this.fb.nonNullable.control(item?.code ?? ''),
      name: this.fb.nonNullable.control(item?.name ?? ''),
      unit: this.fb.nonNullable.control(item?.unit ?? ''),
      quantity: this.fb.control<number | null>(item?.quantity ?? 1, [
        Validators.required,
        Validators.min(1),
        CustomValidators.integer
      ]),
      unitPrice: this.fb.control<number | null>(item?.unitPrice ?? null, [
        Validators.required,
        Validators.min(0)
      ])
    });
  }
}
