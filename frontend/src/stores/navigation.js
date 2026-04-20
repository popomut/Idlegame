import { writable, derived } from 'svelte/store';

// History is a stack of page names. The top of the stack is the current page.
const historyStack = writable(['home']);

export const currentPage = derived(historyStack, function (stack) {
  return stack[stack.length - 1];
});

export const canGoBack = derived(historyStack, function (stack) {
  return stack.length > 1;
});

export function navigateTo(page) {
  historyStack.update(function (stack) {
    // Avoid pushing the same page twice in a row
    if (stack[stack.length - 1] === page) {
      return stack;
    }
    return [...stack, page];
  });
}

export function goBack() {
  historyStack.update(function (stack) {
    if (stack.length <= 1) {
      return stack;
    }
    return stack.slice(0, stack.length - 1);
  });
}

export function goHome() {
  historyStack.set(['home']);
}
