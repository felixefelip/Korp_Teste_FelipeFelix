import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { Router } from '@angular/router';

import { ProductService } from '../../../inventory/products/product.service';
import { FlashService } from '../../../shared/flash/flash.service';
import { ApiFailure, readApiFailure } from '../../../shared/forms/http-errors';
import { InvoiceForm, SAVE_FAILURE } from '../invoice-form/invoice-form';
import { InvoicePayload } from '../invoice.model';
import { InvoiceService } from '../invoice.service';

const PRODUCTS_FAILURE = 'Não foi possível carregar os produtos. Tente novamente.';

@Component({
  selector: 'app-invoice-new',
  imports: [InvoiceForm],
  templateUrl: './invoice-new.html',
  styleUrl: './invoice-new.scss'
})
export class InvoiceNew {
  private readonly invoiceService = inject(InvoiceService);
  private readonly productService = inject(ProductService);
  private readonly router = inject(Router);
  private readonly flash = inject(FlashService);

  protected readonly products = this.productService.products;
  protected readonly productsFailed = signal(false);
  protected readonly saving = signal(false);
  protected readonly failure = signal<ApiFailure | null>(null);

  constructor() {
    this.loadProducts();
  }

  protected create(invoice: InvoicePayload): void {
    this.saving.set(true);
    this.failure.set(null);

    this.invoiceService.create(invoice).subscribe({
      next: () => {
        this.saving.set(false);
        this.router.navigate(['/billing/invoices']);
      },
      error: (response: HttpErrorResponse) => {
        this.saving.set(false);
        this.failure.set(readApiFailure(response, SAVE_FAILURE));
      }
    });
  }

  private loadProducts(): void {
    this.productService.list().subscribe({
      error: () => {
        this.productsFailed.set(true);
        this.flash.error(PRODUCTS_FAILURE);
      }
    });
  }
}
