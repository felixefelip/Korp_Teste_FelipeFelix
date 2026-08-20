import { TestBed } from '@angular/core/testing';

import { ProdutoService } from './produto.service';

describe('ProdutoService', () => {
  let servico: ProdutoService;

  const novoProduto = {
    codigo: 'PRD-0100',
    descricao: 'Cadeira de escritório',
    unidade: 'UN',
    precoUnitario: 750.5,
    estoque: 8
  };

  beforeEach(() => {
    TestBed.configureTestingModule({});
    servico = TestBed.inject(ProdutoService);
  });

  describe('proximoCodigo', () => {
    it('sugere o sequencial seguinte ao maior código existente', () => {
      expect(servico.proximoCodigo()).toBe('PRD-0006');
    });

    it('avança a sugestão após um cadastro no padrão', () => {
      servico.cadastrar({ ...novoProduto, codigo: 'PRD-0009' });
      expect(servico.proximoCodigo()).toBe('PRD-0010');
    });

    it('ignora códigos fora do padrão PRD-0000', () => {
      servico.cadastrar({ ...novoProduto, codigo: 'ABC-9999' });
      expect(servico.proximoCodigo()).toBe('PRD-0006');
    });
  });

  describe('cadastrar', () => {
    it('acrescenta o produto à listagem', () => {
      const antes = servico.produtos().length;
      servico.cadastrar(novoProduto);

      expect(servico.produtos()).toHaveLength(antes + 1);
      expect(servico.produtos().at(-1)).toMatchObject(novoProduto);
    });

    it('gera um id único acima do maior existente', () => {
      const maiorAntes = Math.max(...servico.produtos().map((p) => p.id));
      const criado = servico.cadastrar(novoProduto);

      expect(criado.id).toBe(maiorAntes + 1);
    });

    it('normaliza o código para maiúsculas sem espaços', () => {
      const criado = servico.cadastrar({ ...novoProduto, codigo: '  prd-0100 ' });
      expect(criado.codigo).toBe('PRD-0100');
    });

    it('aceita cadastrar dois produtos com o mesmo código', () => {
      const primeiro = servico.cadastrar(novoProduto);
      const segundo = servico.cadastrar(novoProduto);

      expect(segundo.codigo).toBe(primeiro.codigo);
      expect(segundo.id).not.toBe(primeiro.id);
    });

    it('não altera o array anterior da listagem', () => {
      const listaAnterior = servico.produtos();
      servico.cadastrar(novoProduto);

      expect(servico.produtos()).not.toBe(listaAnterior);
      expect(listaAnterior).not.toContainEqual(
        expect.objectContaining({ codigo: 'PRD-0100' })
      );
    });
  });
});
