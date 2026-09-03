import { CurrencyPipe } from '@angular/common';
import { Component, input, output, signal } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';

import { CatalogProduct } from '../catalog.model';
import { CustomFormValidation } from '../../../shared/forms/custom-form-validation';
import {
  ItemGroup,
  fillGroupFromProduct,
  itemICMS,
  itemIPI,
  itemTotal
} from '../invoice-items-form/item-group';

@Component({
  selector: 'app-invoice-item-dialog',
  imports: [CurrencyPipe, ReactiveFormsModule],
  templateUrl: './invoice-item-dialog.html',
  styleUrl: './invoice-item-dialog.scss'
})
export class InvoiceItemDialog {
  readonly form = input.required<ItemGroup>();
  readonly products = input<CatalogProduct[]>([]);

  readonly saved = output<void>();
  readonly cancelled = output<void>();

  protected readonly submitted = signal(false);

  protected readonly total = () => itemTotal(this.form());
  protected readonly icms = () => itemICMS(this.form());
  protected readonly ipi = () => itemIPI(this.form());

  protected error(field: string): string | null {
    return CustomFormValidation.fieldErrorFor(this.form(), this.submitted)(field);
  }

  protected fillFromProduct(): void {
    fillGroupFromProduct(this.form(), this.products());
  }

  protected save(): void {
    this.submitted.set(true);

    if (this.form().invalid) {
      this.form().markAllAsTouched();
      return;
    }

    this.saved.emit();
  }

  protected cancel(): void {
    this.cancelled.emit();
  }
}
