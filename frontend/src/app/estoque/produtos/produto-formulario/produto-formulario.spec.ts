import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';

import { Produto } from '../produto.model';
import { ProdutoService } from '../produto.service';
import { ProdutoFormulario } from './produto-formulario';

describe('ProdutoFormulario', () => {
  let fixture: ComponentFixture<ProdutoFormulario>;
  let servico: {
    proximoCodigo: ReturnType<typeof vi.fn>;
    listar: ReturnType<typeof vi.fn>;
    cadastrar: ReturnType<typeof vi.fn>;
  };
  let navegar: ReturnType<typeof vi.spyOn>;

  const elemento = () => fixture.nativeElement as HTMLElement;

  const campo = <T extends HTMLElement>(id: string) =>
    elemento().querySelector<T>(`#${id}`)!;

  const texto = (el: Element | null | undefined) =>
    (el?.textContent ?? '').replace(/[\u00A0\u202F]/g, ' ').trim();

  const preencher = async (id: string, valor: string) => {
    const input = campo<HTMLInputElement | HTMLSelectElement>(id);
    input.value = valor;
    input.dispatchEvent(new Event('input'));
    input.dispatchEvent(new Event('change'));
    input.dispatchEvent(new Event('blur'));
    await fixture.whenStable();
  };

  const enviar = async () => {
    elemento()
      .querySelector('form')!
      .dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    await fixture.whenStable();
  };

  const erros = () =>
    Array.from(elemento().querySelectorAll('.campo-erro')).map(texto);

  const erroDe = (id: string) =>
    texto(campo(id).parentElement?.querySelector('.campo-erro'));

  const falha = () => texto(elemento().querySelector('.form__falha'));

  const recusarCom = (corpo: unknown, status = 400) =>
    servico.cadastrar.mockReturnValue(
      throwError(() => new HttpErrorResponse({ status, error: corpo }))
    );

  const preencherFormularioValido = async () => {
    await preencher('name', 'Cadeira de escritório');
    await preencher('unit', 'CX');
    await preencher('price', '750.5');
    await preencher('stock', '8');
  };

  const preencherMinimo = async () => {
    await preencher('name', 'Cadeira de escritório');
    await preencher('price', '750.5');
  };

  beforeEach(async () => {
    servico = {
      proximoCodigo: vi.fn().mockReturnValue('PRD-0006'),
      listar: vi.fn(() => of([] as Produto[])),
      cadastrar: vi.fn((dados: Omit<Produto, 'id'>) => of({ ...dados, id: 6 }))
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
      expect(campo<HTMLInputElement>('code').value).toBe('PRD-0006');
    });

    it('carrega a listagem para poder sugerir o código', () => {
      expect(servico.listar).toHaveBeenCalled();
    });

    it('começa com o estoque inicial zerado', () => {
      expect(campo<HTMLInputElement>('stock').value).toBe('0');
    });

    it('começa com a primeira unidade selecionada', () => {
      expect(campo<HTMLSelectElement>('unit').value).toBe('UN');
    });

    it('lista as unidades de medida disponíveis', () => {
      const opcoes = Array.from(campo<HTMLSelectElement>('unit').options).map(
        (opcao) => opcao.value
      );
      expect(opcoes).toEqual(['UN', 'CX', 'PC', 'KG', 'L', 'M']);
    });

    it('não mostra nenhum erro antes de qualquer interação', () => {
      expect(erros()).toEqual([]);
      expect(falha()).toBe('');
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
      expect(erroDe('name')).toBe('Campo obrigatório.');
      expect(erroDe('price')).toBe('Campo obrigatório.');
      expect(erroDe('stock')).toBe('');
    });

    it('cobra o estoque quando o campo é esvaziado', async () => {
      await preencher('stock', '');
      expect(erroDe('stock')).toBe('Campo obrigatório.');
    });

    it('exige descrição com pelo menos 3 caracteres', async () => {
      await preencher('name', 'ab');
      expect(erroDe('name')).toBe('Informe pelo menos 3 caracteres.');
    });

    it('recusa código com caracteres inválidos', async () => {
      await preencher('code', 'PRD 001/A');
      expect(erroDe('code')).toBe('Use apenas letras, números e hífen.');
    });

    it('aceita código já usado por outro produto', async () => {
      await preencherFormularioValido();
      await preencher('code', 'PRD-0001');
      await enviar();

      expect(erroDe('code')).toBe('');
      expect(servico.cadastrar).toHaveBeenCalledWith(
        expect.objectContaining({ code: 'PRD-0001' })
      );
    });

    it('recusa preço negativo', async () => {
      await preencher('price', '-1');
      expect(erroDe('price')).toBe('O valor não pode ser negativo.');
    });

    it('recusa estoque negativo', async () => {
      await preencher('stock', '-5');
      expect(erroDe('stock')).toBe('O valor não pode ser negativo.');
    });

    it('recusa estoque fracionado', async () => {
      await preencher('stock', '2.5');
      expect(erroDe('stock')).toBe('Informe um número inteiro.');
    });

    it('aceita preço e estoque zerados', async () => {
      await preencherFormularioValido();
      await preencher('price', '0');
      await preencher('stock', '0');
      await enviar();

      expect(servico.cadastrar).toHaveBeenCalledWith(
        expect.objectContaining({ price: 0, stock: 0 })
      );
    });

    it('marca o campo inválido visualmente', async () => {
      await enviar();
      expect(campo('name').classList).toContain('campo--erro');
    });

    it('limpa o erro assim que o campo é corrigido', async () => {
      await enviar();
      expect(erroDe('name')).toBe('Campo obrigatório.');

      await preencher('name', 'Cadeira de escritório');
      expect(erroDe('name')).toBe('');
    });
  });

  describe('cadastro', () => {
    it('envia os dados preenchidos para o serviço', async () => {
      await preencherFormularioValido();
      await enviar();

      expect(servico.cadastrar).toHaveBeenCalledWith({
        code: 'PRD-0006',
        name: 'Cadeira de escritório',
        unit: 'CX',
        price: 750.5,
        stock: 8
      });
    });

    it('remove espaços sobrando da descrição', async () => {
      await preencherFormularioValido();
      await preencher('name', '   Mesa de reunião   ');
      await enviar();

      expect(servico.cadastrar).toHaveBeenCalledWith(
        expect.objectContaining({ name: 'Mesa de reunião' })
      );
    });

    it('salva com o estoque padrão quando o campo não é alterado', async () => {
      await preencherMinimo();
      await enviar();

      expect(servico.cadastrar).toHaveBeenCalledWith(
        expect.objectContaining({ stock: 0 })
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

  describe('recusa da API', () => {
    it('mostra no campo o erro que o servidor apontou', async () => {
      recusarCom({ errors: { code: 'obrigatorio' } });

      await preencherFormularioValido();
      await enviar();

      expect(erroDe('code')).toBe('Campo obrigatório.');
      expect(navegar).not.toHaveBeenCalled();
    });

    it('traduz as demais mensagens do servidor', async () => {
      recusarCom({ errors: { price: 'nao pode ser menor que 0' } });

      await preencherFormularioValido();
      await enviar();

      expect(erroDe('price')).toBe('O valor não pode ser negativo.');
    });

    it('exibe mensagem desconhecida do servidor com a inicial maiúscula', async () => {
      recusarCom({ errors: { unit: 'precisa ser um de: UN, CX' } });

      await preencherFormularioValido();
      await enviar();

      expect(erroDe('unit')).toBe('Precisa ser um de: UN, CX.');
    });

    it('some com o erro do servidor assim que o campo é corrigido', async () => {
      recusarCom({ errors: { code: 'obrigatorio' } });

      await preencherFormularioValido();
      await enviar();
      expect(erroDe('code')).toBe('Campo obrigatório.');

      await preencher('code', 'PRD-0042');
      expect(erroDe('code')).toBe('');
    });

    it('mostra um aviso geral quando o corpo não aponta campo', async () => {
      recusarCom({ message: 'o corpo precisa ser um JSON valido' });

      await preencherFormularioValido();
      await enviar();

      expect(falha()).toBe('o corpo precisa ser um JSON valido');
      expect(navegar).not.toHaveBeenCalled();
    });

    it('mostra um aviso geral quando a API está fora do ar', async () => {
      recusarCom(null, 500);

      await preencherFormularioValido();
      await enviar();

      expect(falha()).toBe('Não foi possível salvar o produto. Tente novamente.');
    });

    it('permite corrigir e tentar de novo', async () => {
      recusarCom({ errors: { code: 'obrigatorio' } });

      await preencherFormularioValido();
      await enviar();

      servico.cadastrar.mockReturnValue(of({ id: 6 } as Produto));
      await preencher('code', 'PRD-0042');
      await enviar();

      expect(servico.cadastrar).toHaveBeenCalledTimes(2);
      expect(navegar).toHaveBeenCalledWith(['/estoque/produtos']);
    });
  });
});
