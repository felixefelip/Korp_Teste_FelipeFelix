import { CurrencyPipe } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';

import { ConfirmDialog } from '../../../shared/confirm-dialog/confirm-dialog';
import { FlashService } from '../../../shared/flash/flash.service';
import { MenuButton, MenuItem } from '../../../shared/menu-button/menu-button';

import {
  INVOICE_STATUS_LABELS,
  INVOICE_TYPE_LABELS,
  InvoiceStatus,
  InvoiceType
} from '../invoice.model';
import { Invoice } from '../invoice.model';
import { InvoiceService } from '../invoice.service';

const DELETE_FAILURE = 'Não foi possível excluir a nota fiscal. Tente novamente.';

@Component({
  selector: 'app-invoice-list',
  imports: [ConfirmDialog, CurrencyPipe, MenuButton, RouterLink],
  templateUrl: './invoice-list.html',
  styleUrl: './invoice-list.scss'
})
export class InvoiceList {
  private readonly invoiceService = inject(InvoiceService);
  private readonly flash = inject(FlashService);

  protected readonly filter = signal('');
  protected readonly loading = signal(true);
  protected readonly failed = signal(false);
  protected readonly pendingDeletion = signal<Invoice | null>(null);
  protected readonly deleting = signal(false);

  protected readonly invoices = computed(() => {
    const term = this.filter().trim().toLowerCase();
    const list = this.invoiceService.invoices();

    if (!term) {
      return list;
    }

    return list.filter((invoice) => invoice.number.toLowerCase().includes(term));
  });

  protected readonly rows = computed(() =>
    this.invoices().map((invoice) => ({
      invoice,
      actions: this.actionsFor(invoice)
    }))
  );

  constructor() {
    this.load();
  }

  private actionsFor(invoice: Invoice): MenuItem[] {
    const actions: MenuItem[] = [
      { label: 'Editar', link: ['/billing/invoices', invoice.id, 'edit'] }
    ];

    if (!this.closed(invoice)) {
      actions.push({ label: 'Excluir', action: () => this.askToDelete(invoice) });
    }

    return actions;
  }

  private closed(invoice: Invoice): boolean {
    return invoice.status === 'CLOSED';
  }

  protected askToDelete(invoice: Invoice): void {
    this.pendingDeletion.set(invoice);
  }

  protected cancelDeletion(): void {
    if (this.deleting()) {
      return;
    }

    this.pendingDeletion.set(null);
  }

  protected confirmDeletion(): void {
    const invoice = this.pendingDeletion();

    if (!invoice || this.deleting()) {
      return;
    }

    this.deleting.set(true);

    this.invoiceService.remove(invoice.id).subscribe({
      next: () => {
        this.deleting.set(false);
        this.pendingDeletion.set(null);
        this.flash.success(`Nota fiscal ${invoice.number} excluída.`);
      },
      error: (response: HttpErrorResponse) => {
        this.deleting.set(false);
        this.pendingDeletion.set(null);
        this.flash.error(response.error?.message ?? DELETE_FAILURE);
      }
    });
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
