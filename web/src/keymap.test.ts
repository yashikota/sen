import { describe, expect, it } from 'vite-plus/test';
import { actionFromKeyboard, isTypingTarget } from './keymap.ts';

function el(tagName: string): EventTarget {
  return { tagName, isContentEditable: false } as unknown as EventTarget;
}

describe('actionFromKeyboard', () => {
  it('opens the palette with mod+k even in an input', () => {
    expect(
      actionFromKeyboard({
        key: 'k',
        metaKey: true,
        ctrlKey: false,
        target: el('INPUT'),
      }),
    ).toBe('palette');
  });

  it('ignores c while typing', () => {
    const input = el('TEXTAREA');
    expect(isTypingTarget(input)).toBe(true);
    expect(
      actionFromKeyboard({
        key: 'c',
        metaKey: false,
        ctrlKey: false,
        target: input,
      }),
    ).toBeNull();
  });

  it('maps list motion and create keys', () => {
    const body = el('BODY');
    expect(
      actionFromKeyboard({
        key: 'j',
        metaKey: false,
        ctrlKey: false,
        target: body,
      }),
    ).toBe('move-down');
    expect(
      actionFromKeyboard({
        key: 'c',
        metaKey: false,
        ctrlKey: false,
        target: body,
      }),
    ).toBe('new-issue');
    expect(
      actionFromKeyboard({
        key: '1',
        metaKey: false,
        ctrlKey: false,
        target: body,
      }),
    ).toBe('priority-1');
    expect(
      actionFromKeyboard({
        key: '?',
        metaKey: false,
        ctrlKey: false,
        target: body,
      }),
    ).toBe('help');
    expect(
      actionFromKeyboard({
        key: '/',
        metaKey: false,
        ctrlKey: false,
        target: body,
      }),
    ).toBe('find');
  });

  it('maps remaining letter and number shortcuts', () => {
    const body = el('BODY');
    expect(
      actionFromKeyboard({
        key: 'p',
        metaKey: false,
        ctrlKey: false,
        target: body,
      }),
    ).toBe('new-page');
    expect(
      actionFromKeyboard({
        key: 's',
        metaKey: false,
        ctrlKey: false,
        target: body,
      }),
    ).toBe('status');
    expect(
      actionFromKeyboard({
        key: 'k',
        metaKey: false,
        ctrlKey: false,
        target: body,
      }),
    ).toBe('move-up');
    expect(
      actionFromKeyboard({
        key: 'Enter',
        metaKey: false,
        ctrlKey: false,
        target: body,
      }),
    ).toBe('open');
    expect(
      actionFromKeyboard({
        key: '2',
        metaKey: false,
        ctrlKey: false,
        target: body,
      }),
    ).toBe('priority-2');
    expect(
      actionFromKeyboard({
        key: '3',
        metaKey: false,
        ctrlKey: false,
        target: body,
      }),
    ).toBe('priority-3');
    expect(
      actionFromKeyboard({
        key: '4',
        metaKey: false,
        ctrlKey: false,
        target: body,
      }),
    ).toBe('priority-4');
  });

  it('opens the palette with ctrl+k', () => {
    expect(
      actionFromKeyboard({
        key: 'k',
        metaKey: false,
        ctrlKey: true,
        target: el('BODY'),
      }),
    ).toBe('palette');
  });

  it('ignores alt+k and unknown keys', () => {
    const body = el('BODY');
    expect(
      actionFromKeyboard({
        key: 'k',
        metaKey: false,
        ctrlKey: false,
        altKey: true,
        target: body,
      }),
    ).toBeNull();
    expect(
      actionFromKeyboard({
        key: 'x',
        metaKey: false,
        ctrlKey: false,
        target: body,
      }),
    ).toBeNull();
  });

  it('lets Escape through while typing and ignores other keys', () => {
    const input = el('INPUT');
    expect(
      actionFromKeyboard({
        key: 'Escape',
        metaKey: false,
        ctrlKey: false,
        target: input,
      }),
    ).toBe('escape');
    expect(
      actionFromKeyboard({
        key: '/',
        metaKey: false,
        ctrlKey: false,
        target: input,
      }),
    ).toBeNull();
  });

  it('treats contenteditable as a typing target', () => {
    const editor = { tagName: 'DIV', isContentEditable: true } as unknown as EventTarget;
    expect(isTypingTarget(editor)).toBe(true);
    expect(
      actionFromKeyboard({
        key: 'c',
        metaKey: false,
        ctrlKey: false,
        target: editor,
      }),
    ).toBeNull();
  });

  it('ignores create while a modifier is held', () => {
    expect(
      actionFromKeyboard({
        key: 'c',
        metaKey: true,
        ctrlKey: false,
        target: el('BODY'),
      }),
    ).toBeNull();
  });
});
