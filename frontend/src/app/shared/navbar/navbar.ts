import { Component } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';

@Component({
  selector: 'app-navbar',
  imports: [RouterLink, RouterLinkActive],
  templateUrl: './navbar.html',
  styleUrl: './navbar.scss'
})
export class Navbar {
  protected readonly menus = [
    { label: 'Produtos', rota: '/estoque/produtos' },
    { label: 'Notas Fiscais', rota: '/faturamento/notas-fiscais' }
  ];
}
