import { CurrencyPipe } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';

import { InvoiceActions } from '../invoice-actions/invoice-actions';

import {
  INVOICE_STATUS_LABELS,
  INVOICE_TYPE_LABELS,
  InvoiceStatus,
  InvoiceType
} from '../invoice.model';
import { InvoiceService } from '../invoice.service';


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

    return list.filter((invoice) => invoice.number.toLowerCase().includes(term));
  });


  constructor() {
    this.load();
  }










  protected typeLabel(type: InvoiceType): string {
    return INVOICE_TYPE_LABELS[type] ?? type;
  }

  protected statusLabel(status: InvoiceStatus): string {
    return INVOICE_STATUS_LABELS[status] ?? status;
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
