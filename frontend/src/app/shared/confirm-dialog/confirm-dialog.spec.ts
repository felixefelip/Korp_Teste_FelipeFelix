import { Component, signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ConfirmDialog } from './confirm-dialog';

@Component({
  imports: [ConfirmDialog],
  template: `
    <app-confirm-dialog
      title="Excluir produto"
      [busy]="busy()"
      (confirmed)="events.push('confirmed')"
      (cancelled)="events.push('cancelled')"
    >
      O produto <strong>PRD-0001</strong> será apagado.
    </app-confirm-dialog>
  `
})
class Host {
  readonly busy = signal(false);
  readonly events: string[] = [];
}

describe('ConfirmDialog', () => {
  let fixture: ComponentFixture<Host>;

  const element = () => fixture.nativeElement as HTMLElement;

  const text = (el: Element | null | undefined) =>
    (el?.textContent ?? '').replace(/\s+/g, ' ').trim();

  const button = (label: string) =>
    Array.from(element().querySelectorAll<HTMLButtonElement>('.confirm__actions .btn')).find(
      (item) => text(item).startsWith(label)
    )!;

  const click = async (el: Element) => {
    (el as HTMLElement).click();
    await fixture.whenStable();
  };

  const events = () => fixture.componentInstance.events;

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [Host] }).compileComponents();

    fixture = TestBed.createComponent(Host);
    await fixture.whenStable();
  });

  it('shows the title it was given', () => {
    expect(text(element().querySelector('.confirm__title'))).toBe('Excluir produto');
  });

  it('projects the body the caller wrote', () => {
    expect(text(element().querySelector('.confirm__body'))).toBe(
      'O produto PRD-0001 será apagado.'
    );
  });

  it('announces itself as a modal dialog', () => {
    const dialog = element().querySelector('.confirm')!;

    expect(dialog.getAttribute('role')).toBe('dialog');
    expect(dialog.getAttribute('aria-modal')).toBe('true');
    expect(dialog.getAttribute('aria-labelledby')).toBe('confirm-title');
    expect(element().querySelector('#confirm-title')).not.toBeNull();
  });

  it('emits on confirm', async () => {
    await click(button('Excluir'));

    expect(events()).toEqual(['confirmed']);
  });

  it('emits on cancel', async () => {
    await click(button('Cancelar'));

    expect(events()).toEqual(['cancelled']);
  });

  it('cancels when the backdrop is clicked', async () => {
    await click(element().querySelector('.overlay')!);

    expect(events()).toEqual(['cancelled']);
  });

  it('stays put when the dialog itself is clicked', async () => {
    await click(element().querySelector('.confirm')!);

    expect(events()).toEqual([]);
  });

  describe('while busy', () => {
    beforeEach(async () => {
      fixture.componentInstance.busy.set(true);
      await fixture.whenStable();
    });

    it('says so on the confirm button', () => {
      expect(text(button('Excluindo'))).toBe('Excluindo…');
    });

    it('disables both buttons', () => {
      expect(button('Excluindo').disabled).toBe(true);
      expect(button('Cancelar').disabled).toBe(true);
    });

    it('emits nothing on a second confirm', async () => {
      await click(button('Excluindo'));

      expect(events()).toEqual([]);
    });

    it('emits nothing when the backdrop is clicked', async () => {
      await click(element().querySelector('.overlay')!);

      expect(events()).toEqual([]);
    });
  });
});
