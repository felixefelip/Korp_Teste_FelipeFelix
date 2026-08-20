import { Injectable, signal } from '@angular/core';

export type FlashType = 'error' | 'success';

export interface FlashMessage {
  id: number;
  type: FlashType;
  text: string;
}

export const FLASH_DISMISS_AFTER = 6000;

@Injectable({ providedIn: 'root' })
export class FlashService {
  private readonly _messages = signal<FlashMessage[]>([]);

  private lastId = 0;

  readonly messages = this._messages.asReadonly();

  error(text: string): void {
    this.show('error', text);
  }

  success(text: string): void {
    this.show('success', text);
  }

  dismiss(id: number): void {
    this._messages.update((messages) => messages.filter((message) => message.id !== id));
  }

  private show(type: FlashType, text: string): void {
    const id = ++this.lastId;

    this._messages.update((messages) => [...messages, { id, type, text }]);

    setTimeout(() => this.dismiss(id), FLASH_DISMISS_AFTER);
  }
}
