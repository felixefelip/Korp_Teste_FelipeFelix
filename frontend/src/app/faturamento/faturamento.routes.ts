import { Routes } from '@angular/router';

export const faturamentoRoutes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'notas-fiscais' },
  {
    path: 'notas-fiscais',
    title: 'Notas Fiscais | Korp ERP',
    loadComponent: () =>
      import('./notas-fiscais/nota-fiscal-lista/nota-fiscal-lista').then(
        (m) => m.NotaFiscalLista
      )
  }
];
