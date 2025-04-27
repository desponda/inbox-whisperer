// jest-dom adds custom jest matchers for asserting on DOM nodes.
// allows you to do things like:
// expect(element).toHaveTextContent(/react/i)
// learn more: https://github.com/testing-library/jest-dom
import '@testing-library/jest-dom';

// Mock TextEncoder/TextDecoder
class TextEncoderMock {
  encode(str: string): Uint8Array {
    return new Uint8Array([]);
  }
}

class TextDecoderMock {
  decode(arr: Uint8Array): string {
    return '';
  }
}

global.TextEncoder = TextEncoderMock as any;
global.TextDecoder = TextDecoderMock as any;

// Mock fetch API
global.fetch = jest.fn(() =>
  Promise.resolve({
    ok: true,
    json: () => Promise.resolve({ id: 'test-user-id' }),
  } as Response)
);

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(),
    removeListener: jest.fn(),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
});
