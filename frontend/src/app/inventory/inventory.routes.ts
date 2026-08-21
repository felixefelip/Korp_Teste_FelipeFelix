import { Routes } from '@angular/router';

export const inventoryRoutes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'products' },
  {
    path: 'products',
    title: 'Produtos | Korp ERP',
    loadComponent: () =>
      import('./products/product-list/product-list').then((m) => m.ProductList)
  },
  {
    path: 'products/new',
    title: 'Cadastrar produto | Korp ERP',
    loadComponent: () =>
      import('./products/product-new/product-new').then((m) => m.ProductNew)
  },
  {
    path: 'products/:id/edit',
    title: 'Editar produto | Korp ERP',
    loadComponent: () =>
      import('./products/product-edit/product-edit').then((m) => m.ProductEdit)
  },
  {
    path: 'products/:id/movements',
    title: 'Movimentações | Korp ERP',
    loadComponent: () =>
      import('./movements/movement-list/movement-list').then((m) => m.MovementList)
  },
  {
    path: 'products/:id/movements/new',
    title: 'Nova movimentação | Korp ERP',
    loadComponent: () =>
      import('./movements/movement-new/movement-new').then((m) => m.MovementNew)
  },
  {
    path: 'products/:id/movements/:movementId/edit',
    title: 'Editar movimentação | Korp ERP',
    loadComponent: () =>
      import('./movements/movement-edit/movement-edit').then((m) => m.MovementEdit)
  }
];
