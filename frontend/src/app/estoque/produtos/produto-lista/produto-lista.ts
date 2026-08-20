import { Component, computed, inject, signal } from '@angular/core';
import { CurrencyPipe } from '@angular/common';
import { RouterLink } from '@angular/router';
import { ProdutoService } from '../produto.service';

@Component({
  selector: 'app-produto-lista',
  imports: [CurrencyPipe, RouterLink],
  templateUrl: './produto-lista.html',
  styleUrl: './produto-lista.scss'
})
export class ProdutoLista {
  private readonly produtoService = inject(ProdutoService);

  protected readonly filtro = signal('');
  protected readonly carregando = signal(true);
  protected readonly erro = signal(false);

  protected readonly produtos = computed(() => {
    const termo = this.filtro().trim().toLowerCase();
    const lista = this.produtoService.produtos();

    if (!termo) {
      return lista;
    }

    return lista.filter(
      (produto) =>
        produto.name.toLowerCase().includes(termo) ||
        produto.code.toLowerCase().includes(termo)
    );
  });

  constructor() {
    this.carregar();
  }

  protected carregar(): void {
    this.carregando.set(true);
    this.erro.set(false);

    this.produtoService.listar().subscribe({
      next: () => this.carregando.set(false),
      error: () => {
        this.carregando.set(false);
        this.erro.set(true);
      }
    });
  }

  protected aoFiltrar(evento: Event): void {
    this.filtro.set((evento.target as HTMLInputElement).value);
  }
}
