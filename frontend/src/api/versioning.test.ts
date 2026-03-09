import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  createVersionHeaders,
  addVersionQuery,
  addVersionPath,
  parseVersionResponse,
  isVersionSupported,
  isVersionDeprecated,
  getVersionDeprecationInfo,
  validateVersion,
  getAPIFeatures,
  VersionError,
  DeprecatedVersionError,
  SunsetVersionError,
  VersionStrategy,
  DEFAULT_VERSION_CONFIG,
  CURRENT_API_VERSION,
} from './versions';

describe('API Versioning Utilities', () => {
  describe('createVersionHeaders', () => {
    it('creates Accept header by default', () => {
      const headers = createVersionHeaders('v2');
      expect(headers).toEqual({
        Accept: 'application/vnd.openagent.v2+json',
      });
    });

    it('creates API-Version header when specified', () => {
      const headers = createVersionHeaders('v1', VersionStrategy.API_HEADER);
      expect(headers).toEqual({
        'API-Version': 'v1',
      });
    });

    it('uses current version by default', () => {
      const headers = createVersionHeaders();
      expect(headers).toEqual({
        Accept: `application/vnd.openagent.${CURRENT_API_VERSION}+json`,
      });
    });
  });

  describe('addVersionQuery', () => {
    it('adds version query parameter to URL without existing params', () => {
      const result = addVersionQuery('/api/agents', 'v2');
      expect(result).toBe('/api/agents?api-version=v2');
    });

    it('adds version query parameter to URL with existing params', () => {
      const result = addVersionQuery('/api/agents?limit=10', 'v2');
      expect(result).toBe('/api/agents?limit=10&api-version=v2');
    });
  });

  describe('addVersionPath', () => {
    it('adds version to API path', () => {
      const result = addVersionPath('/api/agents', 'v2');
      expect(result).toBe('/api/v2/agents');
    });

    it('handles root API path', () => {
      const result = addVersionPath('/api/', 'v1');
      expect(result).toBe('/api/v1/');
    });
  });

  describe('parseVersionResponse', () => {
    it('parses version from response headers', () => {
      const headers = {
        'api-version': 'v2',
      };
      const result = parseVersionResponse(headers);
      expect(result).toEqual({
        version: 'v2',
      });
    });

    it('parses deprecation information', () => {
      const headers = {
        'api-version': 'v1',
        'deprecation': 'true',
        'deprecation-date': '2024-01-15T00:00:00Z',
        'sunset': 'Mon, 15 Jul 2024 00:00:00 GMT',
        'link': '</docs/controlplane/api-versioning.md>; rel="deprecation"; type="text/html"',
      };
      const result = parseVersionResponse(headers);
      expect(result).toEqual({
        version: 'v1',
        isDeprecated: true,
        deprecationDate: '2024-01-15T00:00:00Z',
        sunsetDate: 'Mon, 15 Jul 2024 00:00:00 GMT',
        migrationGuide: '/docs/controlplane/api-versioning.md',
      });
    });

    it('uses current version as fallback', () => {
      const headers = {};
      const result = parseVersionResponse(headers);
      expect(result).toEqual({
        version: CURRENT_API_VERSION,
      });
    });
  });

  describe('isVersionSupported', () => {
    it('returns true for supported version', () => {
      expect(isVersionSupported('v1')).toBe(true);
    });

    it('returns false for unsupported version', () => {
      expect(isVersionSupported('v999')).toBe(false);
    });
  });

  describe('isVersionDeprecated', () => {
    const config = {
      ...DEFAULT_VERSION_CONFIG,
      supportedVersions: [
        { version: 'v1', isDefault: true, isDeprecated: false },
        { version: 'v0', isDefault: false, isDeprecated: true },
      ],
    };

    it('returns false for non-deprecated version', () => {
      expect(isVersionDeprecated('v1', config)).toBe(false);
    });

    it('returns true for deprecated version', () => {
      expect(isVersionDeprecated('v0', config)).toBe(true);
    });

    it('returns false for unknown version', () => {
      expect(isVersionDeprecated('v999', config)).toBe(false);
    });
  });

  describe('getVersionDeprecationInfo', () => {
    const config = {
      ...DEFAULT_VERSION_CONFIG,
      supportedVersions: [
        { version: 'v1', isDefault: true, isDeprecated: false },
        {
          version: 'v0',
          isDefault: false,
          isDeprecated: true,
          deprecatedAt: '2024-01-15T00:00:00Z',
          sunsetAt: '2024-07-15T00:00:00Z',
        },
      ],
    };

    it('returns null for non-deprecated version', () => {
      expect(getVersionDeprecationInfo('v1', config)).toBeNull();
    });

    it('returns deprecation info for deprecated version', () => {
      const result = getVersionDeprecationInfo('v0', config);
      expect(result).toEqual({
        deprecatedAt: new Date('2024-01-15T00:00:00Z'),
        sunsetAt: new Date('2024-07-15T00:00:00Z'),
      });
    });

    it('handles missing deprecation dates', () => {
      const configWithoutDates = {
        ...DEFAULT_VERSION_CONFIG,
        supportedVersions: [
          { version: 'v0', isDefault: false, isDeprecated: true },
        ],
      };
      const result = getVersionDeprecationInfo('v0', configWithoutDates);
      expect(result).toEqual({
        deprecatedAt: undefined,
        sunsetAt: undefined,
      });
    });
  });

  describe('validateVersion', () => {
    beforeEach(() => {
      vi.useFakeTimers();
      vi.setSystemTime(new Date('2024-06-01T00:00:00Z'));
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it('passes validation for supported non-deprecated version', () => {
      expect(() => validateVersion('v1')).not.toThrow();
    });

    it('throws VersionError for unsupported version', () => {
      expect(() => validateVersion('v999')).toThrow(VersionError);
      try {
        validateVersion('v999');
      } catch (error) {
        expect(error).toBeInstanceOf(VersionError);
        expect((error as VersionError).version).toBe('v999');
        expect((error as VersionError).supportedVersions).toEqual(['v1']);
      }
    });

    it('throws DeprecatedVersionError for deprecated but not sunset version', () => {
      const config = {
        ...DEFAULT_VERSION_CONFIG,
        supportedVersions: [
          { version: 'v1', isDefault: true, isDeprecated: false },
          {
            version: 'v0',
            isDefault: false,
            isDeprecated: true,
            deprecatedAt: '2024-01-15T00:00:00Z',
            sunsetAt: '2024-08-15T00:00:00Z', // Future date
          },
        ],
      };
      
      expect(() => validateVersion('v0', config)).toThrow(DeprecatedVersionError);
      try {
        validateVersion('v0', config);
      } catch (error) {
        expect(error).toBeInstanceOf(DeprecatedVersionError);
        expect((error as DeprecatedVersionError).version).toBe('v0');
        expect((error as DeprecatedVersionError).deprecatedAt).toEqual(
          new Date('2024-01-15T00:00:00Z')
        );
        expect((error as DeprecatedVersionError).sunsetAt).toEqual(
          new Date('2024-08-15T00:00:00Z')
        );
      }
    });

    it('throws SunsetVersionError for sunset version', () => {
      const config = {
        ...DEFAULT_VERSION_CONFIG,
        supportedVersions: [
          { version: 'v1', isDefault: true, isDeprecated: false },
          {
            version: 'v0',
            isDefault: false,
            isDeprecated: true,
            sunsetAt: '2024-05-15T00:00:00Z', // Past date
          },
        ],
      };
      
      expect(() => validateVersion('v0', config)).toThrow(SunsetVersionError);
      try {
        validateVersion('v0', config);
      } catch (error) {
        expect(error).toBeInstanceOf(SunsetVersionError);
        expect((error as SunsetVersionError).version).toBe('v0');
        expect((error as SunsetVersionError).sunsetAt).toEqual(
          new Date('2024-05-15T00:00:00Z')
        );
      }
    });
  });

  describe('getAPIFeatures', () => {
    it('returns all features for v1', () => {
      const features = getAPIFeatures('v1');
      expect(features).toEqual({
        batchEventIngestion: true,
        webhookSupport: true,
        advancedFiltering: true,
        realTimeUpdates: true,
      });
    });

    it('returns default features for unknown version', () => {
      const features = getAPIFeatures('v999');
      expect(features).toEqual({
        batchEventIngestion: true,
        webhookSupport: true,
        advancedFiltering: true,
        realTimeUpdates: true,
      });
    });
  });

  describe('Error classes', () => {
    it('creates VersionError with correct properties', () => {
      const error = new VersionError('Test message', 'v2', ['v1']);
      expect(error.name).toBe('VersionError');
      expect(error.message).toBe('Test message');
      expect(error.version).toBe('v2');
      expect(error.supportedVersions).toEqual(['v1']);
    });

    it('creates DeprecatedVersionError with correct properties', () => {
      const deprecatedAt = new Date('2024-01-15T00:00:00Z');
      const sunsetAt = new Date('2024-07-15T00:00:00Z');
      const error = new DeprecatedVersionError(
        'Test message',
        'v0',
        deprecatedAt,
        sunsetAt,
        '/docs/migration'
      );
      expect(error.name).toBe('DeprecatedVersionError');
      expect(error.version).toBe('v0');
      expect(error.deprecatedAt).toBe(deprecatedAt);
      expect(error.sunsetAt).toBe(sunsetAt);
      expect(error.migrationGuide).toBe('/docs/migration');
    });

    it('creates SunsetVersionError with correct properties', () => {
      const sunsetAt = new Date('2024-05-15T00:00:00Z');
      const error = new SunsetVersionError(
        'Test message',
        'v0',
        sunsetAt,
        '/docs/migration'
      );
      expect(error.name).toBe('SunsetVersionError');
      expect(error.version).toBe('v0');
      expect(error.sunsetAt).toBe(sunsetAt);
      expect(error.migrationGuide).toBe('/docs/migration');
    });
  });
});

describe('Integration Tests', () => {
  it('handles complete version negotiation workflow', () => {
    // Create headers for v2
    const headers = createVersionHeaders('v2');
    expect(headers.Accept).toBe('application/vnd.openagent.v2+json');

    // Simulate response headers with deprecation
    const responseHeaders = {
      'api-version': 'v2',
      'deprecation': 'true',
      'sunset': 'Mon, 15 Jul 2024 00:00:00 GMT',
    };

    const versionResponse = parseVersionResponse(responseHeaders);
    expect(versionResponse.version).toBe('v2');
    expect(versionResponse.isDeprecated).toBe(true);
    expect(versionResponse.sunsetDate).toBe('Mon, 15 Jul 2024 00:00:00 GMT');
  });

  it('provides feature detection based on version', () => {
    const v1Features = getAPIFeatures('v1');
    expect(v1Features.batchEventIngestion).toBe(true);
    expect(v1Features.webhookSupport).toBe(true);
    expect(v1Features.advancedFiltering).toBe(true);
    expect(v1Features.realTimeUpdates).toBe(true);
  });
});