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
      import('./products/product-form/product-form').then((m) => m.ProductForm)
  },
  {
    path: 'products/:id/edit',
    title: 'Editar produto | Korp ERP',
    loadComponent: () =>
      import('./products/product-form/product-form').then((m) => m.ProductForm)
  }
];
