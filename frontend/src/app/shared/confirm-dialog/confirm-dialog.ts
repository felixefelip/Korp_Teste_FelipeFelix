import { Component, input, output } from '@angular/core';

@Component({
  selector: 'app-confirm-dialog',
  templateUrl: './confirm-dialog.html',
  styleUrl: './confirm-dialog.scss'
})
export class ConfirmDialog {
  readonly title = input.required<string>();
  readonly confirmLabel = input('Excluir');
  readonly busyLabel = input('Excluindo…');
  readonly busy = input(false);

  readonly confirmed = output<void>();
  readonly cancelled = output<void>();

  protected confirm(): void {
    if (!this.busy()) {
      this.confirmed.emit();
    }
  }

  protected cancel(): void {
    if (!this.busy()) {
      this.cancelled.emit();
    }
  }
}
