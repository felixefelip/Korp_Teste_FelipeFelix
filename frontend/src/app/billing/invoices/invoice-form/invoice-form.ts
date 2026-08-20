import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';

import { CustomFormValidation } from '../../../shared/forms/custom-form-validation';
import { ApiFailure } from '../../../shared/forms/http-errors';
import {
  INVOICE_STATUSES,
  INVOICE_STATUS_LABELS,
  Invoice,
  InvoiceStatus
} from '../invoice.model';

export const SAVE_FAILURE = 'Não foi possível salvar a nota fiscal. Tente novamente.';

export type InvoicePayload = Omit<Invoice, 'id'>;

@Component({
  selector: 'app-invoice-form',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './invoice-form.html',
  styleUrl: './invoice-form.scss'
})
export class InvoiceForm {
  private readonly fb = inject(FormBuilder);

  readonly value = input<Invoice | null>(null);
  readonly loading = input(false);
  readonly saving = input(false);
  readonly failure = input<ApiFailure | null>(null);
  readonly submitLabel = input('Salvar nota fiscal');

  readonly save = output<InvoicePayload>();

  protected readonly statuses = INVOICE_STATUSES;
  protected readonly submitted = signal(false);
  protected readonly banner = signal<string | null>(null);

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

  constructor() {
    effect(() => {
      const invoice = this.value();

      if (!invoice) {
        return;
      }

      this.form.setValue({ number: invoice.number, status: invoice.status });
    });

    effect(() => {
      const failure = this.failure();

      if (!failure) {
        this.banner.set(null);
        return;
      }

      const applied = CustomFormValidation.applyMessageErrorsToForm(
        this.form,
        failure.fieldErrors
      );

      this.banner.set(applied ? null : failure.message);
    });
  }

  protected statusLabel(status: InvoiceStatus): string {
    return INVOICE_STATUS_LABELS[status] ?? status;
  }

  protected submit(): void {
    if (this.loading()) {
      return;
    }

    this.submitted.set(true);

    if (this.form.invalid || this.saving()) {
      this.form.markAllAsTouched();
      return;
    }

    const { number, status } = this.form.getRawValue();

    this.save.emit({ number: number.trim(), status });
  }
}
