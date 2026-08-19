import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';

import { Produto } from '../produto.model';
import { ProdutoService } from '../produto.service';
import { ProdutoFormulario } from './produto-formulario';

describe('ProdutoFormulario', () => {
  let fixture: ComponentFixture<ProdutoFormulario>;
  let servico: {
    existeCodigo: ReturnType<typeof vi.fn>;
    proximoCodigo: ReturnType<typeof vi.fn>;
    cadastrar: ReturnType<typeof vi.fn>;
  };
  let navegar: ReturnType<typeof vi.spyOn>;

  const elemento = () => fixture.nativeElement as HTMLElement;

  const campo = <T extends HTMLElement>(id: string) =>
    elemento().querySelector<T>(`#${id}`)!;

  const texto = (el: Element | null | undefined) =>
    (el?.textContent ?? '').replace(/ /g, ' ').trim();

  const preencher = async (id: string, valor: string) => {
    const input = campo<HTMLInputElement | HTMLSelectElement>(id);
    input.value = valor;
    input.dispatchEvent(new Event('input'));
    input.dispatchEvent(new Event('change'));
    input.dispatchEvent(new Event('blur'));
    await fixture.whenStable();
  };

  const marcarAtivo = async (marcado: boolean) => {
    const check = elemento().querySelector<HTMLInputElement>('input[type="checkbox"]')!;
    check.checked = marcado;
    check.dispatchEvent(new Event('change'));
    await fixture.whenStable();
  };

  const enviar = async () => {
    elemento().querySelector('form')!.dispatchEvent(
      new Event('submit', { bubbles: true, cancelable: true })
    );
    await fixture.whenStable();
  };

  const erros = () =>
    Array.from(elemento().querySelectorAll('.campo-erro')).map(texto);

  const erroDe = (id: string) =>
    texto(campo(id).parentElement?.querySelector('.campo-erro'));

  const preencherFormularioValido = async () => {
    await preencher('descricao', 'Cadeira de escritório');
    await preencher('unidade', 'CX');
    await preencher('precoUnitario', '750.5');
    await preencher('estoque', '8');
  };

  /** Só o obrigatório: o estoque já vem preenchido com 0. */
  const preencherMinimo = async () => {
    await preencher('descricao', 'Cadeira de escritório');
    await preencher('precoUnitario', '750.5');
  };

  beforeEach(async () => {
    servico = {
      existeCodigo: vi.fn().mockReturnValue(false),
      proximoCodigo: vi.fn().mockReturnValue('PRD-0006'),
      cadastrar: vi.fn((dados: Omit<Produto, 'id'>) => ({ ...dados, id: 6 }))
    };

    await TestBed.configureTestingModule({
      imports: [ProdutoFormulario],
      providers: [provideRouter([]), { provide: ProdutoService, useValue: servico }]
    }).compileComponents();

    navegar = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);

    fixture = TestBed.createComponent(ProdutoFormulario);
    await fixture.whenStable();
  });

  it('deve ser criado', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('estado inicial', () => {
    it('sugere o próximo código disponível', () => {
      expect(servico.proximoCodigo).toHaveBeenCalled();
      expect(campo<HTMLInputElement>('codigo').value).toBe('PRD-0006');
    });

    it('começa com o estoque inicial zerado', () => {
      expect(campo<HTMLInputElement>('estoque').value).toBe('0');
    });

    it('começa com a primeira unidade e o produto ativo', () => {
      expect(campo<HTMLSelectElement>('unidade').value).toBe('UN');
      expect(
        elemento().querySelector<HTMLInputElement>('input[type="checkbox"]')!.checked
      ).toBe(true);
    });

    it('lista as unidades de medida disponíveis', () => {
      const opcoes = Array.from(campo<HTMLSelectElement>('unidade').options).map(
        (opcao) => opcao.value
      );
      expect(opcoes).toEqual(['UN', 'CX', 'PC', 'KG', 'L', 'M']);
    });

    it('não mostra nenhum erro antes de qualquer interação', () => {
      expect(erros()).toEqual([]);
    });

    it('oferece cancelar voltando para a listagem', () => {
      expect(elemento().querySelector('a.btn--ghost')?.getAttribute('href')).toBe(
        '/estoque/produtos'
      );
    });
  });

  describe('validação', () => {
    it('bloqueia o envio e revela os erros dos campos obrigatórios', async () => {
      await enviar();

      expect(servico.cadastrar).not.toHaveBeenCalled();
      expect(navegar).not.toHaveBeenCalled();
      expect(erroDe('descricao')).toBe('Campo obrigatório.');
      expect(erroDe('precoUnitario')).toBe('Campo obrigatório.');
      expect(erroDe('estoque')).toBe('');
    });

    it('cobra o estoque quando o campo é esvaziado', async () => {
      await preencher('estoque', '');
      expect(erroDe('estoque')).toBe('Campo obrigatório.');
    });

    it('exige descrição com pelo menos 3 caracteres', async () => {
      await preencher('descricao', 'ab');
      expect(erroDe('descricao')).toBe('Informe pelo menos 3 caracteres.');
    });

    it('recusa código com caracteres inválidos', async () => {
      await preencher('codigo', 'PRD 001/A');
      expect(erroDe('codigo')).toBe('Use apenas letras, números e hífen.');
    });

    it('recusa código já cadastrado', async () => {
      servico.existeCodigo.mockReturnValue(true);
      await preencher('codigo', 'PRD-0001');

      expect(erroDe('codigo')).toBe('Já existe um produto com este código.');
    });

    it('recusa preço negativo', async () => {
      await preencher('precoUnitario', '-1');
      expect(erroDe('precoUnitario')).toBe('O valor não pode ser negativo.');
    });

    it('recusa estoque negativo', async () => {
      await preencher('estoque', '-5');
      expect(erroDe('estoque')).toBe('O valor não pode ser negativo.');
    });

    it('recusa estoque fracionado', async () => {
      await preencher('estoque', '2.5');
      expect(erroDe('estoque')).toBe('Informe um número inteiro.');
    });

    it('aceita preço e estoque zerados', async () => {
      await preencherFormularioValido();
      await preencher('precoUnitario', '0');
      await preencher('estoque', '0');
      await enviar();

      expect(servico.cadastrar).toHaveBeenCalledWith(
        expect.objectContaining({ precoUnitario: 0, estoque: 0 })
      );
    });

    it('marca o campo inválido visualmente', async () => {
      await enviar();
      expect(campo('descricao').classList).toContain('campo--erro');
    });

    it('limpa o erro assim que o campo é corrigido', async () => {
      await enviar();
      expect(erroDe('descricao')).toBe('Campo obrigatório.');

      await preencher('descricao', 'Cadeira de escritório');
      expect(erroDe('descricao')).toBe('');
    });
  });

  describe('cadastro', () => {
    it('envia os dados preenchidos para o serviço', async () => {
      await preencherFormularioValido();
      await enviar();

      expect(servico.cadastrar).toHaveBeenCalledWith({
        codigo: 'PRD-0006',
        descricao: 'Cadeira de escritório',
        unidade: 'CX',
        precoUnitario: 750.5,
        estoque: 8,
        ativo: true
      });
    });

    it('remove espaços sobrando da descrição', async () => {
      await preencherFormularioValido();
      await preencher('descricao', '   Mesa de reunião   ');
      await enviar();

      expect(servico.cadastrar).toHaveBeenCalledWith(
        expect.objectContaining({ descricao: 'Mesa de reunião' })
      );
    });

    it('respeita o produto marcado como inativo', async () => {
      await preencherFormularioValido();
      await marcarAtivo(false);
      await enviar();

      expect(servico.cadastrar).toHaveBeenCalledWith(
        expect.objectContaining({ ativo: false })
      );
    });

    it('salva com o estoque padrão quando o campo não é alterado', async () => {
      await preencherMinimo();
      await enviar();

      expect(servico.cadastrar).toHaveBeenCalledWith(
        expect.objectContaining({ estoque: 0 })
      );
    });

    it('volta para a listagem depois de salvar', async () => {
      await preencherFormularioValido();
      await enviar();

      expect(navegar).toHaveBeenCalledWith(['/estoque/produtos']);
    });

    it('cadastra uma única vez por envio', async () => {
      await preencherFormularioValido();
      await enviar();

      expect(servico.cadastrar).toHaveBeenCalledTimes(1);
    });
  });
});
