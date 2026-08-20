import { registerLocaleData } from '@angular/common';
import localePt from '@angular/common/locales/pt';
import { LOCALE_ID, signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { Observable, of, tap, throwError } from 'rxjs';

import { Produto } from '../produto.model';
import { ProdutoService } from '../produto.service';
import { ProdutoLista } from './produto-lista';

registerLocaleData(localePt, 'pt-BR');

const PRODUTOS: Produto[] = [
  {
    id: 1,
    code: 'PRD-0001',
    name: 'Notebook Dell Inspiron 15',
    unit: 'UN',
    price: 4299.9,
    stock: 12
  },
  {
    id: 2,
    code: 'PRD-0002',
    name: 'Monitor LG 24" Full HD',
    unit: 'UN',
    price: 899,
    stock: 34
  },
  {
    id: 3,
    code: 'ABC-9999',
    name: 'Papel Sulfite A4',
    unit: 'CX',
    price: 27.4,
    stock: 0
  }
];

describe('ProdutoLista', () => {
  let fixture: ComponentFixture<ProdutoLista>;

  const texto = (el: Element | null | undefined) =>
    (el?.textContent ?? '').replace(/[\u00A0\u202F]/g, ' ').trim();

  const elemento = () => fixture.nativeElement as HTMLElement;

  const linhas = () =>
    Array.from(elemento().querySelectorAll('tbody tr')).filter(
      (linha) => !linha.querySelector('.tabela__vazio')
    );

  const colunas = (linha: Element) =>
    Array.from(linha.querySelectorAll('td')).map(texto);

  const descricoes = () => linhas().map((linha) => colunas(linha)[1]);

  const digitarNoFiltro = async (termo: string) => {
    const campo = elemento().querySelector<HTMLInputElement>('input[type="search"]')!;
    campo.value = termo;
    campo.dispatchEvent(new Event('input'));
    await fixture.whenStable();
  };

  const montar = async (resposta: () => Observable<Produto[]>) => {
    const produtos = signal<Produto[]>([]);

    const servico = {
      produtos,
      listar: vi.fn(() => resposta().pipe(tap((lista) => produtos.set(lista))))
    };

    await TestBed.configureTestingModule({
      imports: [ProdutoLista],
      providers: [
        provideRouter([]),
        { provide: ProdutoService, useValue: servico },
        { provide: LOCALE_ID, useValue: 'pt-BR' }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(ProdutoLista);
    await fixture.whenStable();

    return servico;
  };

  beforeEach(async () => {
    await montar(() => of(PRODUTOS));
  });

  it('deve ser criado', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('listagem', () => {
    it('exibe todos os produtos vindos da API', () => {
      expect(descricoes()).toEqual([
        'Notebook Dell Inspiron 15',
        'Monitor LG 24" Full HD',
        'Papel Sulfite A4'
      ]);
    });

    it('exibe as colunas do produto na ordem esperada', () => {
      expect(colunas(linhas()[0])).toEqual([
        'PRD-0001',
        'Notebook Dell Inspiron 15',
        'UN',
        'R$ 4.299,90',
        '12'
      ]);
    });

    it('formata o preço unitário em reais com duas casas', () => {
      const precos = linhas().map((linha) => colunas(linha)[3]);
      expect(precos).toEqual(['R$ 4.299,90', 'R$ 899,00', 'R$ 27,40']);
    });

    it('informa a quantidade de produtos no subtítulo', () => {
      expect(texto(elemento().querySelector('.pagina__subtitulo'))).toBe(
        '3 produto(s) encontrado(s)'
      );
    });

    it('exibe a ação de cadastrar produto apontando para o formulário', () => {
      const acao = elemento().querySelector('a.btn--primary');

      expect(texto(acao)).toBe('+ Cadastrar produto');
      expect(acao?.getAttribute('href')).toBe('/estoque/produtos/novo');
    });
  });

  describe('carga', () => {
    it('busca a listagem na API ao abrir a tela', () => {
      expect(TestBed.inject(ProdutoService).listar).toHaveBeenCalled();
    });

    it('avisa quando a API não responde', async () => {
      TestBed.resetTestingModule();
      await montar(() => throwError(() => new Error('rede fora')));

      expect(texto(elemento().querySelector('.tabela__erro'))).toContain(
        'Não foi possível carregar os produtos.'
      );
      expect(linhas()).toHaveLength(0);
    });

    it('tenta de novo quando o usuário pede', async () => {
      TestBed.resetTestingModule();

      let falhar = true;
      await montar(() =>
        falhar ? throwError(() => new Error('rede fora')) : of(PRODUTOS)
      );

      falhar = false;
      elemento().querySelector<HTMLButtonElement>('.tabela__erro button')!.click();
      await fixture.whenStable();

      expect(descricoes()).toHaveLength(3);
    });
  });

  describe('filtro', () => {
    it('filtra pela descrição', async () => {
      await digitarNoFiltro('monitor');
      expect(descricoes()).toEqual(['Monitor LG 24" Full HD']);
    });

    it('filtra pelo código', async () => {
      await digitarNoFiltro('ABC-9999');
      expect(descricoes()).toEqual(['Papel Sulfite A4']);
    });

    it('ignora diferença entre maiúsculas e minúsculas', async () => {
      await digitarNoFiltro('nOtEbOoK');
      expect(descricoes()).toEqual(['Notebook Dell Inspiron 15']);
    });

    it('ignora espaços em volta do termo', async () => {
      await digitarNoFiltro('   papel   ');
      expect(descricoes()).toEqual(['Papel Sulfite A4']);
    });

    it('casa com trechos no meio da descrição', async () => {
      await digitarNoFiltro('dell');
      expect(descricoes()).toEqual(['Notebook Dell Inspiron 15']);
    });

    it('retorna vários produtos quando o termo é comum', async () => {
      await digitarNoFiltro('PRD-');
      expect(descricoes()).toEqual([
        'Notebook Dell Inspiron 15',
        'Monitor LG 24" Full HD'
      ]);
    });

    it('atualiza a contagem do subtítulo ao filtrar', async () => {
      await digitarNoFiltro('PRD-');
      expect(texto(elemento().querySelector('.pagina__subtitulo'))).toBe(
        '2 produto(s) encontrado(s)'
      );
    });

    it('exibe estado vazio quando nada corresponde', async () => {
      await digitarNoFiltro('inexistente');

      expect(linhas()).toHaveLength(0);
      expect(texto(elemento().querySelector('.tabela__vazio'))).toBe(
        'Nenhum produto encontrado.'
      );
    });

    it('volta a listar tudo quando o filtro é limpo', async () => {
      await digitarNoFiltro('monitor');
      await digitarNoFiltro('');

      expect(descricoes()).toHaveLength(3);
    });
  });
});
