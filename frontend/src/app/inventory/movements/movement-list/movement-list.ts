import { HttpErrorResponse, HttpStatusCode } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';

import { FlashService } from '../../../shared/flash/flash.service';
import { Product } from '../../products/product.model';
import { ProductService } from '../../products/product.service';
import {
  MOVEMENT_ORIGIN_LABELS,
  MOVEMENT_TYPE_LABELS,
  Movement,
  MovementOrigin,
  MovementType,
  isFromInvoice
} from '../movement.model';
import { MovementService } from '../movement.service';

const NOT_FOUND_FAILURE = 'Produto não encontrado.';

@Component({
  selector: 'app-movement-list',
  imports: [RouterLink],
  templateUrl: './movement-list.html',
  styleUrl: './movement-list.scss'
})
export class MovementList {
  private readonly movementService = inject(MovementService);
  private readonly productService = inject(ProductService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly flash = inject(FlashService);

  protected readonly productId = Number(this.route.snapshot.paramMap.get('id'));

  protected readonly product = signal<Product | null>(null);
  protected readonly loading = signal(true);
  protected readonly failed = signal(false);

  protected readonly movements = this.movementService.movements;

  protected readonly reserved = computed(() =>
    this.movements()
      .filter((movement) => movement.type === 'out' && !movement.confirmed)
      .reduce((total, movement) => total + movement.quantity, 0)
  );

  protected readonly available = computed(
    () => (this.product()?.stock ?? 0) - this.reserved()
  );

  constructor() {
    this.load();
  }

  protected typeLabel(type: MovementType): string {
    return MOVEMENT_TYPE_LABELS[type] ?? type;
  }

  protected originLabel(origin: MovementOrigin): string {
    return MOVEMENT_ORIGIN_LABELS[origin] ?? origin;
  }

  protected editable(movement: Movement): boolean {
    return !isFromInvoice(movement);
  }

  protected load(): void {
    this.loading.set(true);
    this.failed.set(false);

    this.productService.get(this.productId).subscribe({
      next: (product) => {
        this.product.set(product);
        this.loadMovements();
      },
      error: (response: HttpErrorResponse) => {
        if (response.status === HttpStatusCode.NotFound) {
          this.flash.error(NOT_FOUND_FAILURE);
          this.router.navigate(['/inventory/products']);
          return;
        }

        this.loading.set(false);
        this.failed.set(true);
      }
    });
  }

  private loadMovements(): void {
    this.movementService.list(this.productId).subscribe({
      next: () => this.loading.set(false),
      error: () => {
        this.loading.set(false);
        this.failed.set(true);
      }
    });
  }
}
