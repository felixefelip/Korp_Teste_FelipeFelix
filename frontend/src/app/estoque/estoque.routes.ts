import { Routes } from '@angular/router';

export const estoqueRoutes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'produtos' },
  {
    path: 'produtos',
    title: 'Produtos | Korp ERP',
    loadComponent: () =>
      import('./produtos/produto-lista/produto-lista').then(
        (m) => m.ProdutoLista
      )
  },
  {
    path: 'produtos/novo',
    title: 'Cadastrar produto | Korp ERP',
    loadComponent: () =>
      import('./produtos/produto-formulario/produto-formulario').then(
        (m) => m.ProdutoFormulario
      )
  }
];
