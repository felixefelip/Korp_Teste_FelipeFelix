import { Component, input, output, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { UnresolvedDraftItem, unresolvedDraftMessage } from '../invoice.model';

export const DRAFT_FAILURE = 'Não foi possível interpretar o pedido. Tente novamente.';

@Component({
  selector: 'app-invoice-prompt',
  imports: [FormsModule],
  templateUrl: './invoice-prompt.html',
  styleUrl: './invoice-prompt.scss'
})
export class InvoicePrompt {
  readonly generating = input(false);
  readonly failure = input<string | null>(null);
  readonly unresolved = input<UnresolvedDraftItem[]>([]);

  readonly generate = output<string>();

  protected readonly prompt = signal('');

  protected readonly message = unresolvedDraftMessage;

  protected submit(): void {
    const prompt = this.prompt().trim();

    if (!prompt || this.generating()) {
      return;
    }

    this.generate.emit(prompt);
  }
}
