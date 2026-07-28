import { describe, it, expect } from 'vitest';
import { resolveApiBase, toWebSocketBase, type PageLocation } from './ws-url';

const page = (hostname: string, protocol = 'http:'): PageLocation => ({
  hostname,
  protocol,
  origin: `${protocol}//${hostname}`,
});

describe('resolveApiBase', () => {
  // The reported bug: the image is built with the localhost default, so a
  // browser on any other machine tried to reach its own computer.
  it('rewrites a loopback base to the page host when viewed remotely', () => {
    expect(resolveApiBase('http://localhost:8080/api', page('ark.example.com')))
      .toBe('http://ark.example.com:8080/api');
  });

  it('rewrites 127.0.0.1 the same way', () => {
    expect(resolveApiBase('http://127.0.0.1:8080/api', page('192.168.1.50')))
      .toBe('http://192.168.1.50:8080/api');
  });

  it('keeps the configured port when rewriting', () => {
    expect(resolveApiBase('http://localhost:9999/api', page('ark.example.com')))
      .toBe('http://ark.example.com:9999/api');
  });

  // An https page cannot open a ws:// socket, so the scheme has to follow.
  it('follows the page scheme to avoid mixed content', () => {
    expect(resolveApiBase('http://localhost:8080/api', page('ark.example.com', 'https:')))
      .toBe('https://ark.example.com:8080/api');
  });

  it('leaves local development untouched', () => {
    expect(resolveApiBase('http://localhost:8080/api', page('localhost')))
      .toBe('http://localhost:8080/api');
    expect(resolveApiBase('http://localhost:8080/api', page('127.0.0.1')))
      .toBe('http://localhost:8080/api');
  });

  // An operator who set a real host knows better than this heuristic.
  it('never overrides an explicitly configured public host', () => {
    expect(resolveApiBase('https://api.example.com/api', page('ark.example.com')))
      .toBe('https://api.example.com/api');
  });

  it('leaves a relative base alone', () => {
    expect(resolveApiBase('/api', page('ark.example.com'))).toBe('/api');
  });

  // During SSR a loopback base is genuinely correct for container-to-container
  // calls, so there is nothing to reconcile.
  it('leaves the base alone when there is no page (SSR)', () => {
    expect(resolveApiBase('http://localhost:8080/api', null))
      .toBe('http://localhost:8080/api');
  });

  it('returns an unparseable base unchanged rather than inventing one', () => {
    expect(resolveApiBase('not a url', page('ark.example.com'))).toBe('not a url');
  });
});

describe('toWebSocketBase', () => {
  // Regression guard: a single /^https?:\/\// pattern matches the whole
  // "https://" prefix, so a comparison against "https:" is never true and every
  // deployment silently gets plaintext ws://. That bug shipped here once.
  it('upgrades https to wss', () => {
    expect(toWebSocketBase('https://ark.example.com/api')).toBe('wss://ark.example.com');
  });

  it('maps http to ws', () => {
    expect(toWebSocketBase('http://ark.example.com:8080/api')).toBe('ws://ark.example.com:8080');
  });

  it('strips the /api suffix with or without a trailing slash', () => {
    expect(toWebSocketBase('https://x.test/api/')).toBe('wss://x.test');
    expect(toWebSocketBase('https://x.test/api')).toBe('wss://x.test');
  });
});
