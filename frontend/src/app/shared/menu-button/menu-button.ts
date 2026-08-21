import { DOCUMENT } from '@angular/common';
import {
  Component,
  ElementRef,
  computed,
  effect,
  inject,
  input,
  signal,
  viewChild,
  viewChildren
} from '@angular/core';
import { RouterLink } from '@angular/router';

export interface MenuItem {
  label: string;
  link?: unknown[];
  action?: () => void;
}

const ITEM_HEIGHT = 40;
const MENU_PADDING = 8;

@Component({
  selector: 'app-menu-button',
  imports: [RouterLink],
  templateUrl: './menu-button.html',
  styleUrl: './menu-button.scss'
})
export class MenuButton {
  private readonly document = inject(DOCUMENT);
  private readonly host = inject<ElementRef<HTMLElement>>(ElementRef);

  readonly label = input<string | null>(null);
  readonly items = input.required<MenuItem[]>();

  private readonly toggle = viewChild<ElementRef<HTMLButtonElement>>('toggle');
  private readonly menuItems = viewChildren<ElementRef<HTMLElement>>('menuItem');

  private readonly pendingFocus = signal<number | null>(null);

  protected readonly open = signal(false);
  protected readonly activeIndex = signal(-1);
  protected readonly placement = signal<{
    top: number | null;
    bottom: number | null;
    right: number;
  }>({ top: null, bottom: null, right: 0 });

  protected readonly primary = computed(() => this.items()[0] ?? null);
  protected readonly rest = computed(() => this.items().slice(1));

  protected readonly primaryLabel = computed(
    () => this.label() ?? this.primary()?.label ?? ''
  );

  constructor() {
    effect((onCleanup) => {
      if (!this.open()) {
        return;
      }

      const closeOnOutside = (event: Event) => {
        if (!this.host.nativeElement.contains(event.target as Node)) {
          this.close();
        }
      };

      const closeAndReturn = () => this.close();

      this.document.addEventListener('click', closeOnOutside);
      this.document.addEventListener('scroll', closeAndReturn, true);
      window.addEventListener('resize', closeAndReturn);

      onCleanup(() => {
        this.document.removeEventListener('click', closeOnOutside);
        this.document.removeEventListener('scroll', closeAndReturn, true);
        window.removeEventListener('resize', closeAndReturn);
      });
    });

    effect(() => {
      const index = this.pendingFocus();
      const items = this.menuItems();

      if (index === null || items.length === 0) {
        return;
      }

      this.pendingFocus.set(null);
      this.focusItem(index);
    });
  }

  protected toggleMenu(): void {
    this.open() ? this.close() : this.openMenu();
  }

  protected onToggleKeydown(event: KeyboardEvent): void {
    if (event.key === 'ArrowDown') {
      event.preventDefault();
      this.openMenu(0);
      return;
    }

    if (event.key === 'ArrowUp') {
      event.preventDefault();
      this.openMenu(this.rest().length - 1);
      return;
    }

    if (event.key === 'Escape') {
      this.close();
    }
  }

  protected onMenuKeydown(event: KeyboardEvent, index: number): void {
    const last = this.rest().length - 1;

    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault();
        this.focusItem(index === last ? 0 : index + 1);
        break;
      case 'ArrowUp':
        event.preventDefault();
        this.focusItem(index === 0 ? last : index - 1);
        break;
      case 'Home':
        event.preventDefault();
        this.focusItem(0);
        break;
      case 'End':
        event.preventDefault();
        this.focusItem(last);
        break;
      case 'Escape':
      case 'Tab':
        this.close();
        break;
    }
  }

  protected run(item: MenuItem): void {
    this.close();
    item.action?.();
  }

  protected close(): void {
    if (!this.open()) {
      return;
    }

    this.open.set(false);
    this.activeIndex.set(-1);
    this.toggle()?.nativeElement.focus();
  }

  private openMenu(focusIndex: number | null = null): void {
    const toggle = this.toggle();

    if (!toggle) {
      return;
    }

    this.placement.set(this.placeAround(toggle.nativeElement));
    this.open.set(true);
    this.pendingFocus.set(focusIndex);
  }

  private placeAround(toggle: HTMLElement) {
    const rect = toggle.getBoundingClientRect();
    const right = window.innerWidth - rect.right;
    const height = this.rest().length * ITEM_HEIGHT + MENU_PADDING;

    if (rect.bottom + height > window.innerHeight && rect.top > height) {
      return { top: null, bottom: window.innerHeight - rect.top, right };
    }

    return { top: rect.bottom, bottom: null, right };
  }

  private focusItem(index: number): void {
    this.menuItems()[index]?.nativeElement.focus();
    this.activeIndex.set(index);
  }
}
