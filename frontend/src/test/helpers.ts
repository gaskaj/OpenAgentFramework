/**
 * Generates a test email in the format:
 * WebTesting-YYYYMMDD-HHMMSSUTC-GUID@OpenAgentFramework.com
 */
export function generateTestEmail(): string {
  const now = new Date();
  const yyyy = now.getUTCFullYear().toString();
  const mm = (now.getUTCMonth() + 1).toString().padStart(2, '0');
  const dd = now.getUTCDate().toString().padStart(2, '0');
  const hh = now.getUTCHours().toString().padStart(2, '0');
  const min = now.getUTCMinutes().toString().padStart(2, '0');
  const ss = now.getUTCSeconds().toString().padStart(2, '0');
  const guid = crypto.randomUUID();

  return `WebTesting-${yyyy}${mm}${dd}-${hh}${min}${ss}UTC-${guid}@OpenAgentFramework.com`;
}

export const TEST_PASSWORD = 'TestPassword123!';
export const TEST_DISPLAY_NAME = 'Web Test User';
