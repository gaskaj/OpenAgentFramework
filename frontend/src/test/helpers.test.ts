import { describe, it, expect } from 'vitest';
import { generateTestEmail } from './helpers';

describe('generateTestEmail', () => {
  it('generates email in the correct format', () => {
    const email = generateTestEmail();

    expect(email).toMatch(
      /^WebTesting-\d{8}-\d{6}UTC-[0-9a-f-]{36}@OpenAgentFramework\.com$/,
    );
  });

  it('generates unique emails on each call', () => {
    const email1 = generateTestEmail();
    const email2 = generateTestEmail();

    expect(email1).not.toBe(email2);
  });

  it('uses current UTC date components', () => {
    const now = new Date();
    const email = generateTestEmail();

    const yyyy = now.getUTCFullYear().toString();
    const mm = (now.getUTCMonth() + 1).toString().padStart(2, '0');
    const dd = now.getUTCDate().toString().padStart(2, '0');

    expect(email).toContain(`WebTesting-${yyyy}${mm}${dd}-`);
  });
});
