import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'inventory/products' },
  {
    path: 'inventory',
    loadChildren: () =>
      import('./inventory/inventory.routes').then((m) => m.inventoryRoutes)
  },
  {
    path: 'billing',
    loadChildren: () =>
      import('./billing/billing.routes').then((m) => m.billingRoutes)
  },
  { path: '**', redirectTo: 'inventory/products' }
];
