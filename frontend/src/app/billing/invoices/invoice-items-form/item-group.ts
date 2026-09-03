import { FormArray, FormControl, FormGroup, Validators } from '@angular/forms';

import { CatalogProduct } from '../catalog.model';
import { CustomValidators } from '../../../shared/forms/validators';
import { InvoiceItemPayload } from '../invoice.model';

export type ItemGroup = FormGroup<{
  inventoryId: FormControl<number | null>;
  code: FormControl<string>;
  name: FormControl<string>;
  unit: FormControl<string>;
  quantity: FormControl<number | null>;
  unitPrice: FormControl<number | null>;
  icmsRate: FormControl<number | null>;
  ipiRate: FormControl<number | null>;
}>;

function rateControl(rate: number | null | undefined): FormControl<number | null> {
  return new FormControl<number | null>(rate ?? 0, [
    Validators.required,
    Validators.min(0),
    Validators.max(100)
  ]);
}

export function newItemGroup(item?: InvoiceItemPayload): ItemGroup {
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
    ]),
    icmsRate: rateControl(item?.icmsRate),
    ipiRate: rateControl(item?.ipiRate)
  });
}

export function newItemArray(items: InvoiceItemPayload[] = []): FormArray<ItemGroup> {
  return new FormArray(items.map(newItemGroup));
}

export function itemPayloadOf(row: ItemGroup): InvoiceItemPayload {
  const value = row.getRawValue();

  return {
    inventoryId: value.inventoryId!,
    code: value.code,
    name: value.name,
    unit: value.unit,
    quantity: value.quantity!,
    unitPrice: value.unitPrice!,
    icmsRate: value.icmsRate!,
    ipiRate: value.ipiRate!
  };
}

export function fillGroupFromProduct(row: ItemGroup, products: CatalogProduct[]): void {
  const product = products.find(
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

export function itemTotal(row: ItemGroup): number {
  const { quantity, unitPrice } = row.getRawValue();

  return (quantity ?? 0) * (unitPrice ?? 0);
}

export function itemTax(row: ItemGroup, rate: number | null): number {
  return Math.round(itemTotal(row) * (rate ?? 0)) / 100;
}

export function itemICMS(row: ItemGroup): number {
  return itemTax(row, row.getRawValue().icmsRate);
}

export function itemIPI(row: ItemGroup): number {
  return itemTax(row, row.getRawValue().ipiRate);
}
