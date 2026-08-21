import { HttpErrorResponse, HttpStatusCode } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';

import { ProductService } from '../../../inventory/products/product.service';
import { FlashService } from '../../../shared/flash/flash.service';
import { ApiFailure, readApiFailure } from '../../../shared/forms/http-errors';
import { InvoiceForm, SAVE_FAILURE } from '../invoice-form/invoice-form';
import { Invoice, InvoicePayload } from '../invoice.model';
import { InvoiceService } from '../invoice.service';

const NOT_FOUND_FAILURE = 'Nota fiscal não encontrada.';
const LOAD_FAILURE = 'Não foi possível carregar a nota fiscal. Tente novamente.';
const PRODUCTS_FAILURE = 'Não foi possível carregar os produtos. Tente novamente.';

@Component({
  selector: 'app-invoice-edit',
  imports: [InvoiceForm],
  templateUrl: './invoice-edit.html',
  styleUrl: './invoice-edit.scss'
})
export class InvoiceEdit {
  private readonly invoiceService = inject(InvoiceService);
  private readonly productService = inject(ProductService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly flash = inject(FlashService);

  private readonly invoiceId = Number(this.route.snapshot.paramMap.get('id'));

  protected readonly products = this.productService.products;
  protected readonly productsFailed = signal(false);
  protected readonly invoice = signal<Invoice | null>(null);
  protected readonly loading = signal(true);
  protected readonly saving = signal(false);
  protected readonly failure = signal<ApiFailure | null>(null);

  constructor() {
    this.load();
    this.loadProducts();
  }

  protected update(invoice: InvoicePayload): void {
    this.saving.set(true);
    this.failure.set(null);

    this.invoiceService.update(this.invoiceId, invoice).subscribe({
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

  private load(): void {
    this.invoiceService.get(this.invoiceId).subscribe({
      next: (invoice) => {
        this.invoice.set(invoice);
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.flash.error(
          response.status === HttpStatusCode.NotFound ? NOT_FOUND_FAILURE : LOAD_FAILURE
        );
        this.router.navigate(['/billing/invoices']);
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
