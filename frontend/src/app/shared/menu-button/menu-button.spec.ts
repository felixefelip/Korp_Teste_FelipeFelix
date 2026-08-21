import { Component } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { MenuButton, MenuItem } from './menu-button';

@Component({ template: '' })
class Destination {}

const ITEMS: MenuItem[] = [
  { label: 'Editar', link: ['/inventory/products', 7, 'edit'] },
  { label: 'Movimentações', link: ['/inventory/products', 7, 'movements'] }
];

const THREE: MenuItem[] = [...ITEMS, { label: 'Imprimir', link: ['/x'] }];

describe('MenuButton', () => {
  let fixture: ComponentFixture<MenuButton>;

  const element = () => fixture.nativeElement as HTMLElement;

  const text = (el: Element | null | undefined) =>
    (el?.textContent ?? '').replace(/[  ]/g, ' ').trim();

  const action = () =>
    element().querySelector<HTMLAnchorElement>('.menu-button__action')!;

  const trigger = () =>
    element().querySelector<HTMLButtonElement>('.menu-button__toggle')!;

  const menu = () => element().querySelector('.menu-button__menu');

  const items = () =>
    Array.from(element().querySelectorAll<HTMLAnchorElement>('.menu-button__item'));

  const labels = () => items().map(text);

  const focused = () => text(document.activeElement);

  const click = async () => {
    trigger().click();
    await fixture.whenStable();
  };

  const press = async (key: string, on: HTMLElement = trigger()) => {
    on.dispatchEvent(new KeyboardEvent('keydown', { key, bubbles: true }));
    await fixture.whenStable();
  };

  const setInput = async (name: string, value: unknown) => {
    fixture.componentRef.setInput(name, value);
    await fixture.whenStable();
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MenuButton],
      providers: [provideRouter([{ path: '**', component: Destination }])]
    }).compileComponents();

    fixture = TestBed.createComponent(MenuButton);
    fixture.componentRef.setInput('items', ITEMS);
    await fixture.whenStable();
  });

  it('should be created', () => {
    expect(fixture.componentInstance).toBeTruthy();
  });

  describe('closed', () => {
    it('shows the first item as the button itself', () => {
      expect(text(action())).toBe('Editar');
      expect(menu()).toBeNull();
    });

    it('runs the first action instead of only naming it', () => {
      expect(action().getAttribute('href')).toBe('/inventory/products/7/edit');
    });

    it('follows the first item when the list changes', async () => {
      await setInput('items', [{ label: 'Imprimir', link: ['/x'] }, ...ITEMS]);

      expect(text(action())).toBe('Imprimir');
    });

    it('hides the chevron when there is nothing else to offer', async () => {
      await setInput('items', [ITEMS[0]]);

      expect(trigger()).toBeNull();
      expect(text(action())).toBe('Editar');
    });

    it('says it is collapsed and owns a menu', () => {
      expect(trigger().getAttribute('aria-expanded')).toBe('false');
      expect(trigger().getAttribute('aria-haspopup')).toBe('true');
    });

    it('lets the caller override the name', async () => {
      await setInput('label', 'Ações');

      expect(text(action())).toBe('Ações');
    });
  });

  describe('opening with the mouse', () => {
    it('reveals the items the button itself does not cover', async () => {
      await click();

      expect(labels()).toEqual(['Movimentações']);
      expect(trigger().getAttribute('aria-expanded')).toBe('true');
    });

    it('never repeats the first item inside the menu', async () => {
      await click();

      expect(labels()).not.toContain('Editar');
    });

    it('marks each item as a menu item', async () => {
      await click();

      expect(menu()!.getAttribute('role')).toBe('menu');
      expect(items().map((item) => item.getAttribute('role'))).toEqual(['menuitem']);
    });

    it('links each item where it was told', async () => {
      await click();

      expect(items().map((item) => item.getAttribute('href'))).toEqual([
        '/inventory/products/7/movements'
      ]);
    });

    it('closes on a second click', async () => {
      await click();
      await click();

      expect(menu()).toBeNull();
    });

    it('focuses nothing inside, so the mouse user is not dragged', async () => {
      await click();

      expect(document.activeElement).not.toBe(items()[0]);
    });
  });

  describe('opening with the keyboard', () => {
    beforeEach(async () => {
      await setInput('items', THREE);
    });

    it('lands on the first item with the down arrow', async () => {
      await press('ArrowDown');

      expect(focused()).toBe('Movimentações');
    });

    it('lands on the last item with the up arrow', async () => {
      await press('ArrowUp');

      expect(focused()).toBe('Imprimir');
    });
  });

  describe('walking the open menu', () => {
    beforeEach(async () => {
      await setInput('items', THREE);
      await press('ArrowDown');
    });

    it('opens on the first item of the menu, not on the button action', () => {
      expect(labels()).toEqual(['Movimentações', 'Imprimir']);
      expect(focused()).toBe('Movimentações');
    });

    it('goes down one item at a time', async () => {
      await press('ArrowDown', items()[0]);

      expect(focused()).toBe('Imprimir');
    });

    it('wraps around at the bottom', async () => {
      await press('ArrowDown', items()[0]);
      await press('ArrowDown', items()[1]);

      expect(focused()).toBe('Movimentações');
    });

    it('wraps around at the top', async () => {
      await press('ArrowUp', items()[0]);

      expect(focused()).toBe('Imprimir');
    });

    it('jumps to the ends with Home and End', async () => {
      await press('End', items()[0]);
      expect(focused()).toBe('Imprimir');

      await press('Home', items()[1]);
      expect(focused()).toBe('Movimentações');
    });

    it('closes on Escape and gives the focus back to the chevron', async () => {
      await press('Escape', items()[0]);

      expect(menu()).toBeNull();
      expect(document.activeElement).toBe(trigger());
    });

    it('closes on Tab, letting the focus move on', async () => {
      await press('Tab', items()[0]);

      expect(menu()).toBeNull();
    });

    it('closes once an item is chosen', async () => {
      items()[0].click();
      await fixture.whenStable();

      expect(menu()).toBeNull();
    });
  });

  describe('closing from outside', () => {
    it('closes on a click anywhere else', async () => {
      await click();

      document.body.click();
      await fixture.whenStable();

      expect(menu()).toBeNull();
    });

    it('stays open for a click inside its own menu', async () => {
      await click();

      menu()!.dispatchEvent(new MouseEvent('click', { bubbles: true }));
      await fixture.whenStable();

      expect(menu()).not.toBeNull();
    });

    it('closes when the page scrolls away under it', async () => {
      await click();

      document.dispatchEvent(new Event('scroll'));
      await fixture.whenStable();

      expect(menu()).toBeNull();
    });

    it('closes when the window is resized', async () => {
      await click();

      window.dispatchEvent(new Event('resize'));
      await fixture.whenStable();

      expect(menu()).toBeNull();
    });
  });

  describe('items that act instead of navigating', () => {
    let ran: string[];

    beforeEach(async () => {
      ran = [];

      await setInput('items', [
        ITEMS[0],
        { label: 'Excluir', action: () => ran.push('Excluir') }
      ]);
    });

    it('renders them as buttons, not links', async () => {
      await click();

      const item = items()[0] as unknown as HTMLElement;

      expect(item.tagName).toBe('BUTTON');
      expect(item.getAttribute('role')).toBe('menuitem');
      expect(item.getAttribute('href')).toBeNull();
    });

    it('runs the action when chosen', async () => {
      await click();

      (items()[0] as unknown as HTMLElement).click();
      await fixture.whenStable();

      expect(ran).toEqual(['Excluir']);
    });

    it('closes the menu before the action runs', async () => {
      await click();

      (items()[0] as unknown as HTMLElement).click();
      await fixture.whenStable();

      expect(menu()).toBeNull();
    });

    it('looks like any other item in the menu', async () => {
      await click();

      expect(items()[0].className).toBe('menu-button__item');
    });

    it('reaches them by keyboard like any other item', async () => {
      await press('ArrowDown');

      expect(focused()).toBe('Excluir');
    });
  });

  describe('a first item that acts instead of navigating', () => {
    it('renders the button itself as a button', async () => {
      const ran: string[] = [];

      await setInput('items', [
        { label: 'Excluir', action: () => ran.push('Excluir') },
        ITEMS[1]
      ]);

      expect(action().tagName).toBe('BUTTON');
      expect(text(action())).toBe('Excluir');

      (action() as unknown as HTMLElement).click();
      await fixture.whenStable();

      expect(ran).toEqual(['Excluir']);
    });
  });
});
