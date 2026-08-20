import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';

import { CustomFormValidation } from '../../../shared/forms/custom-form-validation';
import { isClientError } from '../../../shared/forms/http-errors';
import {
  INVOICE_STATUSES,
  INVOICE_STATUS_LABELS,
  InvoiceStatus
} from '../invoice.model';
import { InvoiceService } from '../invoice.service';

const GENERIC_FAILURE = 'Não foi possível salvar a nota fiscal. Tente novamente.';

@Component({
  selector: 'app-invoice-form',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './invoice-form.html',
  styleUrl: './invoice-form.scss'
})
export class InvoiceForm {
  private readonly fb = inject(FormBuilder);
  private readonly invoiceService = inject(InvoiceService);
  private readonly router = inject(Router);

  protected readonly statuses = INVOICE_STATUSES;
  protected readonly submitted = signal(false);
  protected readonly saving = signal(false);
  protected readonly failure = signal<string | null>(null);

  protected readonly form = this.fb.group({
    number: this.fb.nonNullable.control('', [
      Validators.required,
      Validators.maxLength(30)
    ]),
    status: this.fb.nonNullable.control<InvoiceStatus>(
      INVOICE_STATUSES[0],
      Validators.required
    )
  });

  protected readonly fieldError = CustomFormValidation.fieldErrorFor(
    this.form,
    this.submitted
  );

  protected statusLabel(status: InvoiceStatus): string {
    return INVOICE_STATUS_LABELS[status] ?? status;
  }

  protected save(): void {
    this.submitted.set(true);
    this.failure.set(null);

    if (this.form.invalid || this.saving()) {
      this.form.markAllAsTouched();
      return;
    }

    const { number, status } = this.form.getRawValue();

    this.saving.set(true);

    this.invoiceService.create({ number: number.trim(), status }).subscribe({
      next: () => {
        this.saving.set(false);
        this.router.navigate(['/billing/invoices']);
      },
      error: (response: HttpErrorResponse) => {
        this.saving.set(false);
        this.handleServerFailure(response);
      }
    });
  }

  private handleServerFailure(response: HttpErrorResponse): void {
    if (!isClientError(response.status)) {
      this.failure.set(GENERIC_FAILURE);
      return;
    }

    const errors = response.error?.errors;

    if (CustomFormValidation.applyMessageErrorsToForm(this.form, errors)) {
      this.failure.set(null);
      return;
    }

    const message = response.error?.message;

    this.failure.set(typeof message === 'string' && message ? message : GENERIC_FAILURE);
  }
}
