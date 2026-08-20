import { TestBed } from '@angular/core/testing';

import { FLASH_DISMISS_AFTER, FlashService } from './flash.service';

describe('FlashService', () => {
  let service: FlashService;

  const texts = () => service.messages().map((message) => message.text);

  beforeEach(() => {
    vi.useFakeTimers();
    TestBed.configureTestingModule({});
    service = TestBed.inject(FlashService);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('starts with nothing to show', () => {
    expect(service.messages()).toEqual([]);
  });

  it('publishes the error it was given', () => {
    service.error('Produto não encontrado.');

    expect(texts()).toEqual(['Produto não encontrado.']);
    expect(service.messages()[0].type).toBe('error');
  });

  it('publishes a success apart from an error', () => {
    service.success('Produto salvo.');

    expect(service.messages()[0].type).toBe('success');
  });

  it('stacks the messages in the order they arrived', () => {
    service.error('Primeiro');
    service.success('Segundo');

    expect(texts()).toEqual(['Primeiro', 'Segundo']);
  });

  it('gives each message its own id', () => {
    service.error('Primeiro');
    service.error('Primeiro');

    const [first, second] = service.messages();

    expect(first.id).not.toBe(second.id);
  });

  it('drops the message the user dismissed', () => {
    service.error('Primeiro');
    service.error('Segundo');

    service.dismiss(service.messages()[0].id);

    expect(texts()).toEqual(['Segundo']);
  });

  it('ignores a dismiss of something already gone', () => {
    service.error('Primeiro');
    const { id } = service.messages()[0];

    service.dismiss(id);
    service.dismiss(id);

    expect(service.messages()).toEqual([]);
  });

  it('withdraws the message on its own after the timeout', () => {
    service.error('Produto não encontrado.');

    vi.advanceTimersByTime(FLASH_DISMISS_AFTER - 1);
    expect(texts()).toEqual(['Produto não encontrado.']);

    vi.advanceTimersByTime(1);
    expect(service.messages()).toEqual([]);
  });

  it('counts the timeout of each message from its own arrival', () => {
    service.error('Primeiro');

    vi.advanceTimersByTime(FLASH_DISMISS_AFTER / 2);
    service.error('Segundo');

    vi.advanceTimersByTime(FLASH_DISMISS_AFTER / 2);
    expect(texts()).toEqual(['Segundo']);

    vi.advanceTimersByTime(FLASH_DISMISS_AFTER / 2);
    expect(service.messages()).toEqual([]);
  });

  it('does not break when the user dismissed before the timeout', () => {
    service.error('Primeiro');
    service.dismiss(service.messages()[0].id);

    service.error('Segundo');
    vi.advanceTimersByTime(FLASH_DISMISS_AFTER);

    expect(service.messages()).toEqual([]);
  });
});
