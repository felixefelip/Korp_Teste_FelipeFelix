import { HttpErrorResponse, HttpStatusCode } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { Router } from '@angular/router';

import { CatalogService } from '../catalog.service';
import { FlashService } from '../../../shared/flash/flash.service';
import { ApiFailure, readApiFailure } from '../../../shared/forms/http-errors';
import { InvoiceForm, SAVE_FAILURE } from '../invoice-form/invoice-form';
import { DRAFT_FAILURE, InvoicePrompt } from '../invoice-prompt/invoice-prompt';
import { InvoiceDocument, InvoiceDraft, InvoicePayload } from '../invoice.model';
import { InvoiceService } from '../invoice.service';

const PRODUCTS_FAILURE = 'Não foi possível carregar os produtos. Tente novamente.';

@Component({
  selector: 'app-invoice-new',
  imports: [InvoiceForm, InvoicePrompt],
  templateUrl: './invoice-new.html',
  styleUrl: './invoice-new.scss'
})
export class InvoiceNew {
  private readonly invoiceService = inject(InvoiceService);
  private readonly catalogService = inject(CatalogService);
  private readonly router = inject(Router);
  private readonly flash = inject(FlashService);

  protected readonly products = this.catalogService.products;
  protected readonly productsFailed = signal(false);
  protected readonly saving = signal(false);
  protected readonly failure = signal<ApiFailure | null>(null);

  protected readonly suggestion = signal<InvoiceDocument | null>(null);
  protected readonly draft = signal<InvoiceDraft | null>(null);
  protected readonly drafting = signal(false);
  protected readonly draftFailure = signal<string | null>(null);

  constructor() {
    this.loadProducts();
    this.suggestDocument();
  }

  protected suggestDocument(series?: number): void {
    this.invoiceService.nextDocument(series).subscribe({
      next: (document) => this.suggestion.set(document),
      error: () => this.suggestion.set(null)
    });
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

  protected generate(prompt: string): void {
    this.drafting.set(true);
    this.draftFailure.set(null);

    this.invoiceService.draft(prompt).subscribe({
      next: (draft) => {
        this.drafting.set(false);
        this.draft.set(draft);
      },
      error: (response: HttpErrorResponse) => {
        this.drafting.set(false);
        this.draft.set(null);
        this.draftFailure.set(draftFailureMessage(response));
      }
    });
  }

  private loadProducts(): void {
    this.catalogService.list().subscribe({
      error: () => {
        this.productsFailed.set(true);
        this.flash.error(PRODUCTS_FAILURE);
      }
    });
  }
}

function draftFailureMessage(response: HttpErrorResponse): string {
  const message = response.error?.message;

  if (
    response.status === HttpStatusCode.ServiceUnavailable &&
    typeof message === 'string' &&
    message
  ) {
    return message;
  }

  return DRAFT_FAILURE;
}
