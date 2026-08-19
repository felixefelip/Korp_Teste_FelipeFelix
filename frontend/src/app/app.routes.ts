import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'estoque/produtos' },
  {
    path: 'estoque',
    loadChildren: () =>
      import('./estoque/estoque.routes').then((m) => m.estoqueRoutes)
  },
  {
    path: 'faturamento',
    loadChildren: () =>
      import('./faturamento/faturamento.routes').then(
        (m) => m.faturamentoRoutes
      )
  },
  { path: '**', redirectTo: 'estoque/produtos' }
];
