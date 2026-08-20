import { Component, inject, signal } from '@angular/core';
import {
  AbstractControl,
  FormBuilder,
  ReactiveFormsModule,
  ValidationErrors,
  Validators
} from '@angular/forms';
import { Router, RouterLink } from '@angular/router';

import { ProdutoService } from '../produto.service';

export const UNIDADES = ['UN', 'CX', 'PC', 'KG', 'L', 'M'];

/** Aceita apenas valores inteiros (usado na quantidade em estoque). */
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

  protected readonly formulario = this.fb.group({
    codigo: this.fb.nonNullable.control(this.produtoService.proximoCodigo(), [
      Validators.required,
      Validators.pattern(/^[A-Za-z0-9-]+$/)
    ]),
    descricao: this.fb.nonNullable.control('', [
      Validators.required,
      Validators.minLength(3),
      Validators.maxLength(120)
    ]),
    unidade: this.fb.nonNullable.control(UNIDADES[0], Validators.required),
    precoUnitario: this.fb.control<number | null>(null, [
      Validators.required,
      Validators.min(0)
    ]),
    estoque: this.fb.control<number | null>(0, [
      Validators.required,
      Validators.min(0),
      inteiro
    ])
  });

  /** Mensagem de erro do campo, ou null enquanto não houver o que mostrar. */
  protected erro(campo: string): string | null {
    const controle = this.formulario.get(campo);

    if (!controle || controle.valid || !(controle.touched || this.enviado())) {
      return null;
    }

    const chave = Object.keys(controle.errors ?? {})[0];
    return MENSAGENS[chave] ?? 'Valor inválido.';
  }

  protected salvar(): void {
    this.enviado.set(true);

    if (this.formulario.invalid) {
      this.formulario.markAllAsTouched();
      return;
    }

    const { codigo, descricao, unidade, precoUnitario, estoque } =
      this.formulario.getRawValue();

    this.produtoService.cadastrar({
      codigo,
      descricao: descricao.trim(),
      unidade,
      precoUnitario: precoUnitario!,
      estoque: estoque!
    });

    this.router.navigate(['/estoque/produtos']);
  }
}
