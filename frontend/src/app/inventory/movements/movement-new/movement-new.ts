import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';

import { ApiFailure, readApiFailure } from '../../../shared/forms/http-errors';
import { MovementForm, SAVE_FAILURE } from '../movement-form/movement-form';
import { MovementPayload } from '../movement.model';
import { MovementService } from '../movement.service';

@Component({
  selector: 'app-movement-new',
  imports: [MovementForm],
  templateUrl: './movement-new.html',
  styleUrl: './movement-new.scss'
})
export class MovementNew {
  private readonly movementService = inject(MovementService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);

  protected readonly productId = Number(this.route.snapshot.paramMap.get('id'));

  protected readonly saving = signal(false);
  protected readonly failure = signal<ApiFailure | null>(null);

  protected readonly backLink = ['/inventory/products', this.productId, 'movements'];

  protected create(movement: MovementPayload): void {
    this.saving.set(true);
    this.failure.set(null);

    this.movementService.create(this.productId, movement).subscribe({
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
}
