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

  protected readonly produtos = computed(() => {
    const termo = this.filtro().trim().toLowerCase();
    const lista = this.produtoService.produtos();

    if (!termo) {
      return lista;
    }

    return lista.filter(
      (produto) =>
        produto.descricao.toLowerCase().includes(termo) ||
        produto.codigo.toLowerCase().includes(termo)
    );
  });

  protected aoFiltrar(evento: Event): void {
    this.filtro.set((evento.target as HTMLInputElement).value);
  }
}
