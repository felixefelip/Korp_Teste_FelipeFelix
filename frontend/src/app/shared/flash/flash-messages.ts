import { Component, inject } from '@angular/core';

import { FlashService } from './flash.service';

@Component({
  selector: 'app-flash-messages',
  templateUrl: './flash-messages.html',
  styleUrl: './flash-messages.scss'
})
export class FlashMessages {
  private readonly flashService = inject(FlashService);

  protected readonly messages = this.flashService.messages;

  protected dismiss(id: number): void {
    this.flashService.dismiss(id);
  }
}
