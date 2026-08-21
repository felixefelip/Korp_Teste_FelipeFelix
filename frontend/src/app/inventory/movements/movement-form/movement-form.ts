import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';

import { CustomFormValidation } from '../../../shared/forms/custom-form-validation';
import { ApiFailure } from '../../../shared/forms/http-errors';
import { CustomValidators } from '../../../shared/forms/validators';
import {
  MOVEMENT_TYPE_LABELS,
  Movement,
  MovementPayload,
  MovementType
} from '../movement.model';

export const MOVEMENT_TYPES: MovementType[] = ['in', 'out'];

export const SAVE_FAILURE = 'Não foi possível salvar a movimentação. Tente novamente.';

@Component({
  selector: 'app-movement-form',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './movement-form.html',
  styleUrl: './movement-form.scss'
})
export class MovementForm {
  private readonly fb = inject(FormBuilder);

  readonly value = input<Movement | null>(null);
  readonly backLink = input.required<unknown[]>();
  readonly loading = input(false);
  readonly saving = input(false);
  readonly failure = input<ApiFailure | null>(null);
  readonly submitLabel = input('Salvar movimentação');

  readonly save = output<MovementPayload>();

  protected readonly types = MOVEMENT_TYPES;
  protected readonly typeLabels = MOVEMENT_TYPE_LABELS;
  protected readonly submitted = signal(false);
  protected readonly banner = signal<string | null>(null);

  protected readonly form = this.fb.group({
    type: this.fb.nonNullable.control<MovementType>('in', Validators.required),
    quantity: this.fb.control<number | null>(null, [
      Validators.required,
      Validators.min(1),
      CustomValidators.integer
    ]),
    confirmed: this.fb.nonNullable.control(false)
  });

  protected readonly fieldError = CustomFormValidation.fieldErrorFor(
    this.form,
    this.submitted
  );

  constructor() {
    effect(() => {
      const movement = this.value();

      if (!movement) {
        return;
      }

      this.form.setValue({
        type: movement.type,
        quantity: movement.quantity,
        confirmed: movement.confirmed
      });
    });

    effect(() => {
      const failure = this.failure();

      if (!failure) {
        this.banner.set(null);
        return;
      }

      const applied = CustomFormValidation.applyMessageErrorsToForm(
        this.form,
        failure.fieldErrors
      );

      this.banner.set(applied ? null : failure.message);
    });
  }

  protected submit(): void {
    if (this.loading()) {
      return;
    }

    this.submitted.set(true);

    if (this.form.invalid || this.saving()) {
      this.form.markAllAsTouched();
      return;
    }

    const { type, quantity, confirmed } = this.form.getRawValue();

    this.save.emit({ type, quantity: quantity!, confirmed });
  }
}
