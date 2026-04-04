import { describe, expect, it } from 'vitest';

import { buildUrl, normalize } from './api';

describe('api URL helpers', () => {
  describe('normalize', () => {
    it('removes trailing slashes', () => {
      expect(normalize('http://localhost:8080///')).toBe('http://localhost:8080');
      expect(normalize('https://api.example.com/')).toBe('https://api.example.com');
    });

    it('keeps URL unchanged when it has no trailing slash', () => {
      expect(normalize('http://localhost:8080')).toBe('http://localhost:8080');
    });
  });

  describe('buildUrl', () => {
    const baseUrl = 'http://127.0.0.1:8080';

    it('returns absolute URLs unchanged', () => {
      expect(buildUrl('https://example.com/photos', baseUrl)).toBe('https://example.com/photos');
      expect(buildUrl('http://example.com/devices', baseUrl)).toBe('http://example.com/devices');
    });

    it('builds URL for relative path without leading slash', () => {
      expect(buildUrl('photos', baseUrl)).toBe('http://127.0.0.1:8080/photos');
    });

    it('builds URL for relative path with leading slash', () => {
      expect(buildUrl('/devices', baseUrl)).toBe('http://127.0.0.1:8080/devices');
    });
  });
});
