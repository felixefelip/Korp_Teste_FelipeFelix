import { DOCUMENT } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, input, signal } from '@angular/core';

import { ConfirmDialog } from '../../../shared/confirm-dialog/confirm-dialog';
import { FlashService } from '../../../shared/flash/flash.service';
import { MenuButton, MenuItem } from '../../../shared/menu-button/menu-button';
import { Invoice, isProcessing } from '../invoice.model';
import { InvoiceService } from '../invoice.service';

export const DELETE_FAILURE = 'Não foi possível excluir a nota fiscal. Tente novamente.';
export const CLOSE_FAILURE = 'Não foi possível imprimir a nota fiscal. Tente novamente.';
export const REOPEN_FAILURE = 'Não foi possível reabrir a nota fiscal. Tente novamente.';
export const RETRY_FAILURE = 'Não foi possível reenviar a nota fiscal. Tente novamente.';

@Component({
  selector: 'app-invoice-actions',
  imports: [ConfirmDialog, MenuButton],
  templateUrl: './invoice-actions.html',
  styleUrl: './invoice-actions.scss'
})
export class InvoiceActions {
  private readonly invoiceService = inject(InvoiceService);
  private readonly flash = inject(FlashService);
  private readonly document = inject(DOCUMENT);

  readonly invoice = input.required<Invoice>();

  protected readonly confirmingPrint = signal(false);
  protected readonly printing = signal(false);
  protected readonly confirmingDeletion = signal(false);
  protected readonly deleting = signal(false);

  protected readonly actions = computed<MenuItem[]>(() => {
    const invoice = this.invoice();

    if (isProcessing(invoice)) {
      return [{ label: 'Tentar novamente', action: () => this.retry() }];
    }

    if (invoice.status === 'CLOSED') {
      return [
        { label: 'Ver DANFE', action: () => this.viewDanfe() },
        { label: 'Reabrir', action: () => this.reopen() }
      ];
    }

    return [
      { label: 'Editar', link: ['/billing/invoices', invoice.id, 'edit'] },
      { label: 'Imprimir', action: () => this.confirmingPrint.set(true) },
      { label: 'Excluir', action: () => this.confirmingDeletion.set(true) }
    ];
  });

  protected cancelPrint(): void {
    if (!this.printing()) {
      this.confirmingPrint.set(false);
    }
  }

  protected confirmPrint(): void {
    if (this.printing()) {
      return;
    }

    const invoice = this.invoice();

    this.printing.set(true);

    this.invoiceService.close(invoice.id).subscribe({
      next: () => this.settlePrint(`Nota fiscal ${invoice.formattedNumber} fechada.`),
      error: (response: HttpErrorResponse) =>
        this.settlePrint(null, response.error?.message ?? CLOSE_FAILURE)
    });
  }

  protected cancelDeletion(): void {
    if (!this.deleting()) {
      this.confirmingDeletion.set(false);
    }
  }

  protected confirmDeletion(): void {
    if (this.deleting()) {
      return;
    }

    const invoice = this.invoice();

    this.deleting.set(true);

    this.invoiceService.remove(invoice.id).subscribe({
      next: () => this.settleDeletion(`Nota fiscal ${invoice.formattedNumber} excluída.`),
      error: (response: HttpErrorResponse) =>
        this.settleDeletion(null, response.error?.message ?? DELETE_FAILURE)
    });
  }

  private viewDanfe(): void {
    const url = this.invoiceService.danfeUrl(this.invoice().id);

    this.document.defaultView?.open(url, '_blank', 'noopener');
  }

  private retry(): void {
    const invoice = this.invoice();

    this.invoiceService.retry(invoice.id).subscribe({
      next: () => this.flash.success(`Nota fiscal ${invoice.formattedNumber} reenviada.`),
      error: (response: HttpErrorResponse) =>
        this.flash.error(response.error?.message ?? RETRY_FAILURE)
    });
  }

  private reopen(): void {
    const invoice = this.invoice();

    this.invoiceService.reopen(invoice.id).subscribe({
      next: () => this.flash.success(`Nota fiscal ${invoice.formattedNumber} reaberta.`),
      error: (response: HttpErrorResponse) =>
        this.flash.error(response.error?.message ?? REOPEN_FAILURE)
    });
  }

  private settlePrint(success: string | null, failure?: string): void {
    this.printing.set(false);
    this.confirmingPrint.set(false);

    success ? this.flash.success(success) : this.flash.error(failure!);
  }

  private settleDeletion(success: string | null, failure?: string): void {
    this.deleting.set(false);
    this.confirmingDeletion.set(false);

    success ? this.flash.success(success) : this.flash.error(failure!);
  }
}
