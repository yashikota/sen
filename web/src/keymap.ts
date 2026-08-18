export type KeyAction =
  | 'palette'
  | 'new-issue'
  | 'new-page'
  | 'move-down'
  | 'move-up'
  | 'open'
  | 'escape'
  | 'status'
  | 'priority-1'
  | 'priority-2'
  | 'priority-3'
  | 'priority-4';

const TYPING_TAGS = new Set(['INPUT', 'TEXTAREA', 'SELECT']);

export function isTypingTarget(target: EventTarget | null): boolean {
  if (target === null || typeof target !== 'object') {
    return false;
  }
  const el = target as { tagName?: string; isContentEditable?: boolean };
  if (el.isContentEditable) {
    return true;
  }
  return TYPING_TAGS.has(el.tagName ?? '');
}

export function actionFromKeyboard(event: {
  key: string;
  metaKey: boolean;
  ctrlKey: boolean;
  altKey?: boolean;
  target: EventTarget | null;
}): KeyAction | null {
  const mod = event.metaKey || event.ctrlKey;
  if (mod && event.key.toLowerCase() === 'k') {
    return 'palette';
  }
  if (mod || event.altKey) {
    return null;
  }
  if (event.key === 'Escape') {
    return 'escape';
  }
  if (isTypingTarget(event.target)) {
    return null;
  }
  switch (event.key) {
    case 'c':
      return 'new-issue';
    case 'p':
      return 'new-page';
    case 'j':
      return 'move-down';
    case 'k':
      return 'move-up';
    case 'Enter':
      return 'open';
    case 's':
      return 'status';
    case '1':
      return 'priority-1';
    case '2':
      return 'priority-2';
    case '3':
      return 'priority-3';
    case '4':
      return 'priority-4';
    default:
      return null;
  }
}
