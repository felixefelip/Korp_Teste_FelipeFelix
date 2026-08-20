import { isClientError } from './http-errors';

describe('isClientError', () => {
  it('accepts the 4xx range', () => {
    expect(isClientError(400)).toBe(true);
    expect(isClientError(409)).toBe(true);
    expect(isClientError(422)).toBe(true);
    expect(isClientError(499)).toBe(true);
  });

  it('rejects server errors', () => {
    expect(isClientError(500)).toBe(false);
    expect(isClientError(502)).toBe(false);
  });

  it('rejects the status of a request that never reached the server', () => {
    expect(isClientError(0)).toBe(false);
  });
});
