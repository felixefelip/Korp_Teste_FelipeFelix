import { registerLocaleData } from '@angular/common';
import localePt from '@angular/common/locales/pt';
import { LOCALE_ID, signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { Produto } from '../produto.model';
import { ProdutoService } from '../produto.service';
import { ProdutoLista } from './produto-lista';

registerLocaleData(localePt, 'pt-BR');

const PRODUTOS: Produto[] = [
  {
    id: 1,
    codigo: 'PRD-0001',
    descricao: 'Notebook Dell Inspiron 15',
    unidade: 'UN',
    precoUnitario: 4299.9,
    estoque: 12,
    ativo: true
  },
  {
    id: 2,
    codigo: 'PRD-0002',
    descricao: 'Monitor LG 24" Full HD',
    unidade: 'UN',
    precoUnitario: 899,
    estoque: 34,
    ativo: true
  },
  {
    id: 3,
    codigo: 'ABC-9999',
    descricao: 'Papel Sulfite A4',
    unidade: 'CX',
    precoUnitario: 27.4,
    estoque: 0,
    ativo: false
  }
];

describe('ProdutoLista', () => {
  let fixture: ComponentFixture<ProdutoLista>;

  const texto = (el: Element | null | undefined) =>
    (el?.textContent ?? '').replace(/ /g, ' ').trim();

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

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ProdutoLista],
      providers: [
        provideRouter([]),
        { provide: ProdutoService, useValue: { produtos: signal(PRODUTOS) } },
        { provide: LOCALE_ID, useValue: 'pt-BR' }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(ProdutoLista);
    await fixture.whenStable();
  });

  it('deve ser criado', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('listagem', () => {
    it('exibe todos os produtos vindos do serviço', () => {
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
        '12',
        'Ativo'
      ]);
    });

    it('formata o preço unitário em reais com duas casas', () => {
      const precos = linhas().map((linha) => colunas(linha)[3]);
      expect(precos).toEqual(['R$ 4.299,90', 'R$ 899,00', 'R$ 27,40']);
    });

    it('marca a situação de produtos ativos e inativos', () => {
      const tags = Array.from(elemento().querySelectorAll('tbody .tag'));

      expect(tags.map(texto)).toEqual(['Ativo', 'Ativo', 'Inativo']);
      expect(tags[0].classList).toContain('tag--ativo');
      expect(tags[2].classList).toContain('tag--inativo');
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
