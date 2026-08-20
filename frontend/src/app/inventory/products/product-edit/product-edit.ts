import { HttpErrorResponse, HttpStatusCode } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';

import { FlashService } from '../../../shared/flash/flash.service';
import { ApiFailure, readApiFailure } from '../../../shared/forms/http-errors';
import { ProductForm, ProductPayload, SAVE_FAILURE } from '../product-form/product-form';
import { Product } from '../product.model';
import { ProductService } from '../product.service';

const NOT_FOUND_FAILURE = 'Produto não encontrado.';
const LOAD_FAILURE = 'Não foi possível carregar o produto. Tente novamente.';

@Component({
  selector: 'app-product-edit',
  imports: [ProductForm],
  templateUrl: './product-edit.html',
  styleUrl: './product-edit.scss'
})
export class ProductEdit {
  private readonly productService = inject(ProductService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly flash = inject(FlashService);

  private readonly productId = Number(this.route.snapshot.paramMap.get('id'));

  protected readonly product = signal<Product | null>(null);
  protected readonly loading = signal(true);
  protected readonly saving = signal(false);
  protected readonly failure = signal<ApiFailure | null>(null);

  constructor() {
    this.load();
  }

  protected update(product: ProductPayload): void {
    this.saving.set(true);
    this.failure.set(null);

    this.productService.update(this.productId, product).subscribe({
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

  private load(): void {
    this.productService.get(this.productId).subscribe({
      next: (product) => {
        this.product.set(product);
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.flash.error(
          response.status === HttpStatusCode.NotFound ? NOT_FOUND_FAILURE : LOAD_FAILURE
        );
        this.router.navigate(['/inventory/products']);
      }
    });
  }
}
