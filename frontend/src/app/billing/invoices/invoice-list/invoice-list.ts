import { CurrencyPipe } from '@angular/common';
import { Component, computed, effect, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';

import { InvoiceActions } from '../invoice-actions/invoice-actions';

import {
  isProcessing,
  processingStage,
  INVOICE_FAILURE_LABELS,
  INVOICE_STATUS_LABELS,
  INVOICE_TYPE_LABELS,
  PROCESSING_STAGE_MESSAGES,
  Invoice,
  InvoiceStatus,
  InvoiceType,
  ProcessingStage
} from '../invoice.model';
import { InvoiceService } from '../invoice.service';

export const POLL_INTERVAL = 1500;
export const SLOW_POLL_INTERVAL = 10000;
export const CLOCK_INTERVAL = 1000;


@Component({
  selector: 'app-invoice-list',
  imports: [CurrencyPipe, InvoiceActions, RouterLink],
  templateUrl: './invoice-list.html',
  styleUrl: './invoice-list.scss'
})
export class InvoiceList {
  private readonly invoiceService = inject(InvoiceService);

  protected readonly filter = signal('');
  protected readonly loading = signal(true);
  protected readonly failed = signal(false);

  protected readonly invoices = computed(() => {
    const term = this.filter().trim().toLowerCase();
    const list = this.invoiceService.invoices();

    if (!term) {
      return list;
    }

    return list.filter((invoice) => invoice.formattedNumber.toLowerCase().includes(term));
  });


  protected readonly processing = computed(() =>
    this.invoiceService.invoices().some(isProcessing)
  );

  private readonly now = signal(Date.now());

  private readonly pollInterval = computed(() => {
    const waiting = this.invoiceService.invoices().filter(isProcessing);
    const settlingSoon = waiting.some(
      (invoice) => processingStage(invoice, this.now()) === 'normal'
    );

    return settlingSoon ? POLL_INTERVAL : SLOW_POLL_INTERVAL;
  });

  constructor() {
    this.load();

    effect((onCleanup) => {
      if (!this.processing()) {
        return;
      }

      const timer = setInterval(() => this.refresh(), this.pollInterval());

      onCleanup(() => clearInterval(timer));
    });

    effect((onCleanup) => {
      if (!this.processing()) {
        return;
      }

      const timer = setInterval(() => this.now.set(Date.now()), CLOCK_INTERVAL);

      onCleanup(() => clearInterval(timer));
    });
  }

  private refresh(): void {
    this.invoiceService.list().subscribe({ error: () => undefined });
  }

  protected typeLabel(type: InvoiceType): string {
    return INVOICE_TYPE_LABELS[type] ?? type;
  }

  protected statusLabel(status: InvoiceStatus): string {
    return INVOICE_STATUS_LABELS[status] ?? status;
  }

  protected isProcessing(invoice: Invoice): boolean {
    return isProcessing(invoice);
  }

  protected stage(invoice: Invoice): ProcessingStage {
    return processingStage(invoice, this.now());
  }

  protected processingMessage(invoice: Invoice): string {
    return PROCESSING_STAGE_MESSAGES[this.stage(invoice)];
  }

  protected failureLabel(invoice: Invoice): string {
    const reason = invoice.failureReason ?? '';
    const label = INVOICE_FAILURE_LABELS[reason] ?? reason;
    const codes = (invoice.shortages ?? []).map((shortage) => shortage.code);

    return codes.length ? `${label}: ${codes.join(', ')}` : label;
  }

  protected load(): void {
    this.loading.set(true);
    this.failed.set(false);

    this.invoiceService.list().subscribe({
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
