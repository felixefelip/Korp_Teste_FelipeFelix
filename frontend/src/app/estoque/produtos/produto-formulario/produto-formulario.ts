import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import {
  AbstractControl,
  FormBuilder,
  ReactiveFormsModule,
  ValidationErrors,
  Validators
} from '@angular/forms';
import { Router, RouterLink } from '@angular/router';

import { ErrosDoServidor } from '../produto.model';
import { ProdutoService } from '../produto.service';

export const UNIDADES = ['UN', 'CX', 'PC', 'KG', 'L', 'M'];

function inteiro(controle: AbstractControl): ValidationErrors | null {
  const valor = controle.value;
  return valor === null || valor === '' || Number.isInteger(valor)
    ? null
    : { naoInteiro: true };
}

const MENSAGENS: Record<string, string> = {
  required: 'Campo obrigatório.',
  minlength: 'Informe pelo menos 3 caracteres.',
  maxlength: 'Limite de 120 caracteres excedido.',
  pattern: 'Use apenas letras, números e hífen.',
  min: 'O valor não pode ser negativo.',
  naoInteiro: 'Informe um número inteiro.'
};

const MENSAGENS_DO_SERVIDOR: Record<string, string> = {
  obrigatorio: 'Campo obrigatório.',
  'tipo invalido': 'Valor inválido.',
  'nao pode ser menor que 0': 'O valor não pode ser negativo.'
};

const FALHA_GENERICA = 'Não foi possível salvar o produto. Tente novamente.';

@Component({
  selector: 'app-produto-formulario',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './produto-formulario.html',
  styleUrl: './produto-formulario.scss'
})
export class ProdutoFormulario {
  private readonly fb = inject(FormBuilder);
  private readonly produtoService = inject(ProdutoService);
  private readonly router = inject(Router);

  protected readonly unidades = UNIDADES;
  protected readonly enviado = signal(false);
  protected readonly salvando = signal(false);
  protected readonly falha = signal<string | null>(null);

  protected readonly formulario = this.fb.group({
    code: this.fb.nonNullable.control(this.produtoService.proximoCodigo(), [
      Validators.required,
      Validators.pattern(/^[A-Za-z0-9-]+$/)
    ]),
    name: this.fb.nonNullable.control('', [
      Validators.required,
      Validators.minLength(3),
      Validators.maxLength(120)
    ]),
    unit: this.fb.nonNullable.control(UNIDADES[0], Validators.required),
    price: this.fb.control<number | null>(null, [
      Validators.required,
      Validators.min(0)
    ]),
    stock: this.fb.control<number | null>(0, [
      Validators.required,
      Validators.min(0),
      inteiro
    ])
  });

  constructor() {
    this.produtoService.listar().subscribe({
      next: () => {
        const controle = this.formulario.controls.code;

        if (controle.pristine) {
          controle.setValue(this.produtoService.proximoCodigo());
        }
      },
      error: () => {}
    });
  }

  protected erro(campo: string): string | null {
    const controle = this.formulario.get(campo);

    if (!controle || controle.valid || !(controle.touched || this.enviado())) {
      return null;
    }

    const erros = controle.errors ?? {};

    if (typeof erros['servidor'] === 'string') {
      return erros['servidor'];
    }

    return MENSAGENS[Object.keys(erros)[0]] ?? 'Valor inválido.';
  }

  protected salvar(): void {
    this.enviado.set(true);
    this.falha.set(null);

    if (this.formulario.invalid || this.salvando()) {
      this.formulario.markAllAsTouched();
      return;
    }

    const { code, name, unit, price, stock } = this.formulario.getRawValue();

    this.salvando.set(true);

    this.produtoService
      .cadastrar({ code, name: name.trim(), unit, price: price!, stock: stock! })
      .subscribe({
        next: () => {
          this.salvando.set(false);
          this.router.navigate(['/estoque/produtos']);
        },
        error: (resposta: HttpErrorResponse) => {
          this.salvando.set(false);
          this.tratarFalha(resposta);
        }
      });
  }

  private tratarFalha(resposta: HttpErrorResponse): void {
    const erros = resposta.error?.errors as ErrosDoServidor | undefined;

    if (!erros) {
      this.falha.set(resposta.error?.message ?? FALHA_GENERICA);
      return;
    }

    let algumCampoConhecido = false;

    for (const [campo, mensagem] of Object.entries(erros)) {
      const controle = this.formulario.get(campo);

      if (controle) {
        controle.setErrors({ servidor: this.traduzir(mensagem) });
        algumCampoConhecido = true;
      }
    }

    this.falha.set(algumCampoConhecido ? null : FALHA_GENERICA);
  }

  private traduzir(mensagem: string): string {
    return (
      MENSAGENS_DO_SERVIDOR[mensagem] ??
      `${mensagem.charAt(0).toUpperCase()}${mensagem.slice(1)}.`
    );
  }
}
