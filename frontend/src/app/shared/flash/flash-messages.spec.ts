import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FlashMessage, FlashService } from './flash.service';
import { FlashMessages } from './flash-messages';

describe('FlashMessages', () => {
  let fixture: ComponentFixture<FlashMessages>;
  let messages: ReturnType<typeof signal<FlashMessage[]>>;
  let dismiss: ReturnType<typeof vi.fn>;

  const element = () => fixture.nativeElement as HTMLElement;

  const text = (el: Element | null | undefined) => (el?.textContent ?? '').trim();

  const items = () => Array.from(element().querySelectorAll('.flash__item'));

  const shown = () => items().map((item) => text(item.querySelector('.flash__text')));

  const mount = async (initial: FlashMessage[]) => {
    messages = signal<FlashMessage[]>(initial);
    dismiss = vi.fn((id: number) =>
      messages.update((list) => list.filter((message) => message.id !== id))
    );

    await TestBed.configureTestingModule({
      imports: [FlashMessages],
      providers: [
        { provide: FlashService, useValue: { messages, dismiss } }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(FlashMessages);
    await fixture.whenStable();
  };

  it('shows nothing while there is no message', async () => {
    await mount([]);

    expect(element().querySelector('.flash')).toBeNull();
  });

  it('shows the text of every message', async () => {
    await mount([
      { id: 1, type: 'error', text: 'Produto não encontrado.' },
      { id: 2, type: 'success', text: 'Produto salvo.' }
    ]);

    expect(shown()).toEqual(['Produto não encontrado.', 'Produto salvo.']);
  });

  it('separates an error from a success visually', async () => {
    await mount([
      { id: 1, type: 'error', text: 'Produto não encontrado.' },
      { id: 2, type: 'success', text: 'Produto salvo.' }
    ]);

    expect(items()[0].classList).toContain('flash__item--error');
    expect(items()[1].classList).toContain('flash__item--success');
  });

  it('asks the service to drop the message the user closed', async () => {
    await mount([
      { id: 1, type: 'error', text: 'Produto não encontrado.' },
      { id: 2, type: 'success', text: 'Produto salvo.' }
    ]);

    items()[0].querySelector<HTMLButtonElement>('.flash__close')!.click();
    await fixture.whenStable();

    expect(dismiss).toHaveBeenCalledWith(1);
    expect(shown()).toEqual(['Produto salvo.']);
  });

  it('follows the service when a message leaves on its own', async () => {
    await mount([{ id: 1, type: 'error', text: 'Produto não encontrado.' }]);

    messages.set([]);
    await fixture.whenStable();

    expect(element().querySelector('.flash')).toBeNull();
  });

  it('announces the messages to assistive technology', async () => {
    await mount([{ id: 1, type: 'error', text: 'Produto não encontrado.' }]);

    expect(element().querySelector('.flash')?.getAttribute('aria-live')).toBe('polite');
  });
});
