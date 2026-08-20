import { Component, computed, inject, signal } from '@angular/core';
import { CurrencyPipe } from '@angular/common';
import { RouterLink } from '@angular/router';
import { ProductService } from '../product.service';

@Component({
  selector: 'app-product-list',
  imports: [CurrencyPipe, RouterLink],
  templateUrl: './product-list.html',
  styleUrl: './product-list.scss'
})
export class ProductList {
  private readonly productService = inject(ProductService);

  protected readonly filter = signal('');
  protected readonly loading = signal(true);
  protected readonly failed = signal(false);

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

  protected onFilter(event: Event): void {
    this.filter.set((event.target as HTMLInputElement).value);
  }
}
