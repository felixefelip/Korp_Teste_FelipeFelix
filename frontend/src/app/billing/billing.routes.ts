import { Routes } from '@angular/router';

export const billingRoutes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'invoices' },
  {
    path: 'invoices',
    title: 'Notas Fiscais | Korp ERP',
    loadComponent: () =>
      import('./invoices/invoice-list/invoice-list').then((m) => m.InvoiceList)
  }
];
