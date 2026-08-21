import { Component, computed, inject, signal } from '@angular/core';
import { CurrencyPipe } from '@angular/common';
import { RouterLink } from '@angular/router';

import { FlashService } from '../../../shared/flash/flash.service';
import { MenuButton, MenuItem } from '../../../shared/menu-button/menu-button';
import { Product } from '../product.model';
import { ProductService } from '../product.service';

const DELETE_FAILURE = 'Não foi possível excluir o produto. Tente novamente.';

@Component({
  selector: 'app-product-list',
  imports: [CurrencyPipe, MenuButton, RouterLink],
  templateUrl: './product-list.html',
  styleUrl: './product-list.scss'
})
export class ProductList {
  private readonly productService = inject(ProductService);
  private readonly flash = inject(FlashService);

  protected readonly filter = signal('');
  protected readonly loading = signal(true);
  protected readonly failed = signal(false);
  protected readonly pendingDeletion = signal<Product | null>(null);
  protected readonly deleting = signal(false);

  protected readonly products = computed(() => {
    const term = this.filter().trim().toLowerCase();
    const list = this.productService.products();

    if (!term) {
      return list;
    }

    return list.filter(
      (product) =>
        product.name.toLowerCase().includes(term) ||
        product.code.toLowerCase().includes(term)
    );
  });

  protected readonly rows = computed(() =>
    this.products().map((product) => ({
      product,
      actions: [
        { label: 'Editar', link: ['/inventory/products', product.id, 'edit'] },
        { label: 'Movimentações', link: ['/inventory/products', product.id, 'movements'] },
        { label: 'Excluir', action: () => this.askToDelete(product) }
      ] as MenuItem[]
    }))
  );

  constructor() {
    this.load();
  }

  protected load(): void {
    this.loading.set(true);
    this.failed.set(false);

    this.productService.list().subscribe({
      next: () => this.loading.set(false),
      error: () => {
        this.loading.set(false);
        this.failed.set(true);
      }
    });
  }

  protected askToDelete(product: Product): void {
    this.pendingDeletion.set(product);
  }

  protected cancelDeletion(): void {
    if (this.deleting()) {
      return;
    }

    this.pendingDeletion.set(null);
  }

  protected confirmDeletion(): void {
    const product = this.pendingDeletion();

    if (!product || this.deleting()) {
      return;
    }

    this.deleting.set(true);

    this.productService.remove(product.id).subscribe({
      next: () => {
        this.deleting.set(false);
        this.pendingDeletion.set(null);
        this.flash.success(`Produto ${product.code} excluído.`);
      },
      error: () => {
        this.deleting.set(false);
        this.pendingDeletion.set(null);
        this.flash.error(DELETE_FAILURE);
      }
    });
  }

  protected onFilter(event: Event): void {
    this.filter.set((event.target as HTMLInputElement).value);
  }
}
