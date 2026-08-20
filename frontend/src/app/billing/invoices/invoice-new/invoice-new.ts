import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { Router } from '@angular/router';

import { ApiFailure, readApiFailure } from '../../../shared/forms/http-errors';
import { InvoiceForm, InvoicePayload, SAVE_FAILURE } from '../invoice-form/invoice-form';
import { InvoiceService } from '../invoice.service';

@Component({
  selector: 'app-invoice-new',
  imports: [InvoiceForm],
  templateUrl: './invoice-new.html',
  styleUrl: './invoice-new.scss'
})
export class InvoiceNew {
  private readonly invoiceService = inject(InvoiceService);
  private readonly router = inject(Router);

  protected readonly saving = signal(false);
  protected readonly failure = signal<ApiFailure | null>(null);

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
}
