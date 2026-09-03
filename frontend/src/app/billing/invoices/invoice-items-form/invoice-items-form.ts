import { CurrencyPipe } from '@angular/common';
import { Component, input, signal } from '@angular/core';
import { FormArray, ReactiveFormsModule } from '@angular/forms';

import { CatalogProduct } from '../catalog.model';
import { InvoiceShortage } from '../invoice.model';
import { CustomFormValidation } from '../../../shared/forms/custom-form-validation';
import { MenuButton, MenuItem } from '../../../shared/menu-button/menu-button';
import { InvoiceItemDialog } from '../invoice-item-dialog/invoice-item-dialog';
import {
  ItemGroup,
  fillGroupFromProduct,
  itemICMS,
  itemIPI,
  itemPayloadOf,
  itemTotal,
  newItemGroup
} from './item-group';

export type { ItemGroup } from './item-group';
export { newItemGroup, newItemArray } from './item-group';

const HIDDEN_FIELDS = ['icmsRate', 'ipiRate'];

@Component({
  selector: 'app-invoice-items-form',
  imports: [CurrencyPipe, InvoiceItemDialog, MenuButton, ReactiveFormsModule],
  templateUrl: './invoice-items-form.html',
  styleUrl: './invoice-items-form.scss'
})
export class InvoiceItemsForm {
  readonly items = input.required<FormArray<ItemGroup>>();
  readonly products = input<CatalogProduct[]>([]);
  readonly productsFailed = input(false);
  readonly submitted = input(false);
  readonly shortages = input<InvoiceShortage[]>([]);

  private readonly actionsByRow = new WeakMap<ItemGroup, MenuItem[]>();
  private readonly editingRow = signal<ItemGroup | null>(null);

  protected readonly editingForm = signal<ItemGroup | null>(null);

  protected rowActions(row: ItemGroup): MenuItem[] {
    const cached = this.actionsByRow.get(row);

    if (cached) {
      return cached;
    }

    const actions: MenuItem[] = [
      { label: 'Editar', action: () => this.editItem(row) },
      { label: 'Remover', action: () => this.removeItem(row) }
    ];

    this.actionsByRow.set(row, actions);

    return actions;
  }

  protected rowShortage(row: ItemGroup): InvoiceShortage | null {
    const inventoryId = row.controls.inventoryId.value;

    return this.shortages().find((shortage) => shortage.inventoryId === inventoryId) ?? null;
  }

  protected rowError(row: ItemGroup, field: string): string | null {
    return CustomFormValidation.fieldErrorFor(row, this.submitted)(field);
  }

  protected hiddenError(row: ItemGroup): string | null {
    for (const field of HIDDEN_FIELDS) {
      if (this.rowError(row, field)) {
        return 'Impostos inválidos. Abra Editar para corrigir.';
      }
    }

    return null;
  }

  protected addItem(): void {
    this.items().push(newItemGroup());
  }

  protected removeItem(row: ItemGroup): void {
    const index = this.items().controls.indexOf(row);

    if (index >= 0) {
      this.items().removeAt(index);
    }
  }

  protected editItem(row: ItemGroup): void {
    this.editingRow.set(row);
    this.editingForm.set(newItemGroup(itemPayloadOf(row)));
  }

  protected saveEdit(): void {
    const row = this.editingRow();
    const edited = this.editingForm();

    if (row && edited) {
      row.patchValue(edited.getRawValue());
      row.markAsDirty();
    }

    this.closeEdit();
  }

  protected closeEdit(): void {
    this.editingRow.set(null);
    this.editingForm.set(null);
  }

  protected fillFromProduct(row: ItemGroup): void {
    fillGroupFromProduct(row, this.products());
  }

  protected rowTotal(row: ItemGroup): number {
    return itemTotal(row);
  }

  protected productsTotal(): number {
    return this.sum(itemTotal);
  }

  protected totalICMS(): number {
    return this.sum(itemICMS);
  }

  protected totalIPI(): number {
    return this.sum(itemIPI);
  }

  protected total(): number {
    return this.productsTotal() + this.totalIPI();
  }

  private sum(of: (row: ItemGroup) => number): number {
    return this.items().controls.reduce((total, row) => total + of(row), 0);
  }
}
