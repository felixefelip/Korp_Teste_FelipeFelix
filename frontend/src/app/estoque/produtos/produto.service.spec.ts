import { provideHttpClient } from '@angular/common/http';
import {
  HttpTestingController,
  provideHttpClientTesting
} from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { Produto } from './produto.model';
import { ProdutoService } from './produto.service';

describe('ProdutoService', () => {
  let servico: ProdutoService;
  let http: HttpTestingController;

  const novoProduto = {
    code: 'PRD-0100',
    name: 'Cadeira de escritório',
    unit: 'UN',
    price: 750.5,
    stock: 8
  };

  const produtos: Produto[] = [
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
      code: 'PRD-0005',
      name: 'Papel Sulfite A4 75g (resma)',
      unit: 'CX',
      price: 27.4,
      stock: 56
    }
  ];

  const carregar = (lista: Produto[] = produtos) => {
    servico.listar().subscribe();
    http.expectOne('/api/products').flush(lista);
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()]
    });

    servico = TestBed.inject(ProdutoService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  describe('listar', () => {
    it('começa sem produtos até a primeira carga', () => {
      expect(servico.produtos()).toEqual([]);
    });

    it('busca a listagem no endpoint da API', () => {
      servico.listar().subscribe();

      const requisicao = http.expectOne('/api/products');
      expect(requisicao.request.method).toBe('GET');

      requisicao.flush(produtos);
    });

    it('publica no signal o que a API devolveu', () => {
      carregar();
      expect(servico.produtos()).toEqual(produtos);
    });

    it('substitui a lista anterior a cada carga', () => {
      carregar();
      carregar([produtos[0]]);

      expect(servico.produtos()).toEqual([produtos[0]]);
    });

    it('propaga a falha e mantém a lista anterior', () => {
      carregar();

      let falhou = false;
      servico.listar().subscribe({ error: () => (falhou = true) });
      http.expectOne('/api/products').flush(null, { status: 500, statusText: 'Erro' });

      expect(falhou).toBe(true);
      expect(servico.produtos()).toEqual(produtos);
    });
  });

  describe('proximoCodigo', () => {
    it('sugere o sequencial seguinte ao maior código carregado', () => {
      carregar();
      expect(servico.proximoCodigo()).toBe('PRD-0006');
    });

    it('sugere o primeiro código quando ainda não há produtos', () => {
      expect(servico.proximoCodigo()).toBe('PRD-0001');
    });

    it('avança a sugestão após um cadastro no padrão', () => {
      carregar();

      servico.cadastrar({ ...novoProduto, code: 'PRD-0009' }).subscribe();
      http.expectOne('/api/products').flush({ ...novoProduto, id: 3, code: 'PRD-0009' });

      expect(servico.proximoCodigo()).toBe('PRD-0010');
    });

    it('ignora códigos fora do padrão PRD-0000', () => {
      carregar([...produtos, { ...novoProduto, id: 3, code: 'ABC-9999' }]);
      expect(servico.proximoCodigo()).toBe('PRD-0006');
    });
  });

  describe('cadastrar', () => {
    it('envia o produto para a API e devolve o que ela criou', () => {
      const criado = { ...novoProduto, id: 7 };
      let recebido: Produto | undefined;

      servico.cadastrar(novoProduto).subscribe((produto) => (recebido = produto));

      const requisicao = http.expectOne('/api/products');
      expect(requisicao.request.method).toBe('POST');
      expect(requisicao.request.body).toEqual(novoProduto);

      requisicao.flush(criado);

      expect(recebido).toEqual(criado);
    });

    it('acrescenta à listagem o produto devolvido pela API', () => {
      carregar();

      servico.cadastrar(novoProduto).subscribe();
      http.expectOne('/api/products').flush({ ...novoProduto, id: 7, code: 'PRD-0100' });

      expect(servico.produtos()).toHaveLength(3);
      expect(servico.produtos().at(-1)).toMatchObject({ id: 7, code: 'PRD-0100' });
    });

    it('não altera o array anterior da listagem', () => {
      carregar();
      const listaAnterior = servico.produtos();

      servico.cadastrar(novoProduto).subscribe();
      http.expectOne('/api/products').flush({ ...novoProduto, id: 7 });

      expect(servico.produtos()).not.toBe(listaAnterior);
      expect(listaAnterior).toHaveLength(2);
    });

    it('não acrescenta nada quando a API recusa o produto', () => {
      carregar();

      let falhou = false;
      servico.cadastrar(novoProduto).subscribe({ error: () => (falhou = true) });
      http
        .expectOne('/api/products')
        .flush(
          { errors: { code: 'obrigatorio' } },
          { status: 400, statusText: 'Bad Request' }
        );

      expect(falhou).toBe(true);
      expect(servico.produtos()).toEqual(produtos);
    });
  });
});
