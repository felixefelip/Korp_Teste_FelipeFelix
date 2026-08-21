import { HttpErrorResponse, HttpStatusCode } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';

import { FlashService } from '../../../shared/flash/flash.service';
import { ApiFailure, readApiFailure } from '../../../shared/forms/http-errors';
import { MovementForm, SAVE_FAILURE } from '../movement-form/movement-form';
import { Movement, MovementPayload, isFromInvoice } from '../movement.model';
import { MovementService } from '../movement.service';

const NOT_FOUND_FAILURE = 'Movimentação não encontrada.';
const LOAD_FAILURE = 'Não foi possível carregar a movimentação. Tente novamente.';
const FROM_INVOICE_FAILURE =
  'Movimentações geradas por notas fiscais não podem ser alteradas.';

@Component({
  selector: 'app-movement-edit',
  imports: [MovementForm],
  templateUrl: './movement-edit.html',
  styleUrl: './movement-edit.scss'
})
export class MovementEdit {
  private readonly movementService = inject(MovementService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly flash = inject(FlashService);

  private readonly movementId = Number(this.route.snapshot.paramMap.get('movementId'));

  protected readonly productId = Number(this.route.snapshot.paramMap.get('id'));

  protected readonly movement = signal<Movement | null>(null);
  protected readonly loading = signal(true);
  protected readonly saving = signal(false);
  protected readonly failure = signal<ApiFailure | null>(null);

  protected readonly backLink = ['/inventory/products', this.productId, 'movements'];

  constructor() {
    this.load();
  }

  protected update(movement: MovementPayload): void {
    this.saving.set(true);
    this.failure.set(null);

    this.movementService.update(this.productId, this.movementId, movement).subscribe({
      next: () => {
        this.saving.set(false);
        this.router.navigate(this.backLink);
      },
      error: (response: HttpErrorResponse) => {
        this.saving.set(false);
        this.failure.set(readApiFailure(response, SAVE_FAILURE));
      }
    });
  }

  private load(): void {
    this.movementService.get(this.productId, this.movementId).subscribe({
      next: (movement) => {
        if (isFromInvoice(movement)) {
          this.flash.error(FROM_INVOICE_FAILURE);
          this.router.navigate(this.backLink);
          return;
        }

        this.movement.set(movement);
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.flash.error(
          response.status === HttpStatusCode.NotFound ? NOT_FOUND_FAILURE : LOAD_FAILURE
        );
        this.router.navigate(this.backLink);
      }
    });
  }
}
