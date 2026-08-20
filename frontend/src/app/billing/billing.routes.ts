import { Routes } from '@angular/router';

export const billingRoutes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'invoices' },
  {
    path: 'invoices',
    title: 'Notas Fiscais | Korp ERP',
    loadComponent: () =>
      import('./invoices/invoice-list/invoice-list').then((m) => m.InvoiceList)
  },
  {
    path: 'invoices/new',
    title: 'Cadastrar nota fiscal | Korp ERP',
    loadComponent: () =>
      import('./invoices/invoice-form/invoice-form').then((m) => m.InvoiceForm)
  }
];
