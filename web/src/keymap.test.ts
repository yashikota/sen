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
  });
});
