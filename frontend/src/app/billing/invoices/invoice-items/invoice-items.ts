import { CurrencyPipe } from '@angular/common';
import { Component, input } from '@angular/core';
import {
  FormArray,
  FormControl,
  FormGroup,
  ReactiveFormsModule,
  Validators
} from '@angular/forms';

import { Product } from '../../../inventory/products/product.model';
import { CustomFormValidation } from '../../../shared/forms/custom-form-validation';
import { CustomValidators } from '../../../shared/forms/validators';
import { InvoiceItem } from '../invoice.model';

export type ItemGroup = FormGroup<{
  inventoryId: FormControl<number | null>;
  code: FormControl<string>;
  name: FormControl<string>;
  unit: FormControl<string>;
  quantity: FormControl<number | null>;
  unitPrice: FormControl<number | null>;
}>;

export function newItemGroup(item?: InvoiceItem): ItemGroup {
  return new FormGroup({
    inventoryId: new FormControl<number | null>(
      item?.inventoryId ?? null,
      Validators.required
    ),
    code: new FormControl(item?.code ?? '', { nonNullable: true }),
    name: new FormControl(item?.name ?? '', { nonNullable: true }),
    unit: new FormControl(item?.unit ?? '', { nonNullable: true }),
    quantity: new FormControl<number | null>(item?.quantity ?? 1, [
      Validators.required,
      Validators.min(1),
      CustomValidators.integer
    ]),
    unitPrice: new FormControl<number | null>(item?.unitPrice ?? null, [
      Validators.required,
      Validators.min(0)
    ])
  });
}

export function newItemArray(items: InvoiceItem[] = []): FormArray<ItemGroup> {
  return new FormArray(items.map(newItemGroup));
}

@Component({
  selector: 'app-invoice-items',
  imports: [CurrencyPipe, ReactiveFormsModule],
  templateUrl: './invoice-items.html',
  styleUrl: './invoice-items.scss'
})
export class InvoiceItems {
  readonly items = input.required<FormArray<ItemGroup>>();
  readonly products = input<Product[]>([]);
  readonly productsFailed = input(false);
  readonly submitted = input(false);

  protected rowError(row: ItemGroup, field: string): string | null {
    return CustomFormValidation.fieldErrorFor(row, this.submitted)(field);
  }

  protected addItem(): void {
    this.items().push(newItemGroup());
  }

  protected removeItem(index: number): void {
    this.items().removeAt(index);
  }

  protected fillFromProduct(row: ItemGroup): void {
    const product = this.products().find(
      (candidate) => candidate.id === row.controls.inventoryId.value
    );

    if (!product) {
      return;
    }

    row.patchValue({
      code: product.code,
      name: product.name,
      unit: product.unit,
      unitPrice: product.price
    });
  }

  protected rowTotal(row: ItemGroup): number {
    const { quantity, unitPrice } = row.getRawValue();

    return (quantity ?? 0) * (unitPrice ?? 0);
  }

  protected total(): number {
    return this.items().controls.reduce(
      (total, row) => total + this.rowTotal(row),
      0
    );
  }
}
