import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { Router } from '@angular/router';

import { ApiFailure, readApiFailure } from '../../../shared/forms/http-errors';
import { ProductForm, SAVE_FAILURE } from '../product-form/product-form';
import { ProductPayload } from '../product.model';
import { ProductService } from '../product.service';

@Component({
  selector: 'app-product-new',
  imports: [ProductForm],
  templateUrl: './product-new.html',
  styleUrl: './product-new.scss'
})
export class ProductNew {
  private readonly productService = inject(ProductService);
  private readonly router = inject(Router);

  protected readonly saving = signal(false);
  protected readonly failure = signal<ApiFailure | null>(null);

  protected create(product: ProductPayload): void {
    this.saving.set(true);
    this.failure.set(null);

    this.productService.create(product).subscribe({
      next: () => {
        this.saving.set(false);
        this.router.navigate(['/inventory/products']);
      },
      error: (response: HttpErrorResponse) => {
        this.saving.set(false);
        this.failure.set(readApiFailure(response, SAVE_FAILURE));
      }
    });
  }
}
