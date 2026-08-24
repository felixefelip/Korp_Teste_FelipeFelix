import { ComponentFixture, TestBed } from '@angular/core/testing';

import { UnresolvedDraftItem } from '../invoice.model';
import { InvoicePrompt } from './invoice-prompt';

describe('InvoicePrompt', () => {
  let fixture: ComponentFixture<InvoicePrompt>;
  let generated: string[];

  const element = () => fixture.nativeElement as HTMLElement;

  const text = (el: Element | null | undefined) =>
    (el?.textContent ?? '').replace(/[  ]/g, ' ').trim();

  const setInput = async (name: string, value: unknown) => {
    fixture.componentRef.setInput(name, value);
    await fixture.whenStable();
  };

  const field = () => element().querySelector<HTMLTextAreaElement>('.prompt__field')!;

  const button = () =>
    element().querySelector<HTMLButtonElement>('.prompt__actions .btn')!;

  const issues = () =>
    Array.from(element().querySelectorAll('.prompt__issue')).map(text);

  const describePrompt = async (value: string) => {
    field().value = value;
    field().dispatchEvent(new Event('input'));
    await fixture.whenStable();
  };

  const click = async () => {
    button().click();
    await fixture.whenStable();
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [InvoicePrompt]
    }).compileComponents();

    fixture = TestBed.createComponent(InvoicePrompt);

    generated = [];
    fixture.componentInstance.generate.subscribe((prompt) => generated.push(prompt));

    await fixture.whenStable();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  it('emits the trimmed prompt', async () => {
    await describePrompt('  vender 2 cadeiras  ');
    await click();

    expect(generated).toEqual(['vender 2 cadeiras']);
  });

  it('keeps the button off while the prompt is empty', async () => {
    expect(button().disabled).toBe(true);

    await describePrompt('   ');

    expect(button().disabled).toBe(true);
  });

  it('emits nothing while generating', async () => {
    await describePrompt('vender 2 cadeiras');
    await setInput('generating', true);

    await click();

    expect(generated).toEqual([]);
  });

  it('says that it is working', async () => {
    await setInput('generating', true);

    expect(text(button())).toBe('Montando…');
  });

  it('shows the failure it was given', async () => {
    await setInput('failure', 'Não foi possível interpretar o pedido.');

    expect(text(element().querySelector('.form__failure'))).toBe(
      'Não foi possível interpretar o pedido.'
    );
  });

  it('lists one line per unresolved item', async () => {
    const unresolved: UnresolvedDraftItem[] = [
      { text: 'monitor LG', quantity: 2, reason: 'NOT_FOUND', candidates: [] },
      { text: 'cadeira', quantity: 0, reason: 'INVALID_QUANTITY', candidates: [] }
    ];

    await setInput('unresolved', unresolved);

    expect(issues()).toEqual([
      '"monitor LG": não encontrei esse produto no catálogo.',
      '"cadeira": não consegui identificar a quantidade.'
    ]);
  });

  it('shows nothing when there is nothing to report', () => {
    expect(issues()).toEqual([]);
    expect(element().querySelector('.form__failure')).toBeNull();
  });
});
