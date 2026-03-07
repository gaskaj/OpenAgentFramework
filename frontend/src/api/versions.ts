/**
 * API versioning utilities for the frontend client
 */

export interface APIVersion {
  version: string;
  isDefault: boolean;
  isDeprecated: boolean;
  deprecatedAt?: string;
  sunsetAt?: string;
}

export interface VersionConfig {
  defaultVersion: string;
  supportedVersions: APIVersion[];
  deprecationWarning: boolean;
}

export interface VersionResponse {
  version: string;
  isDeprecated?: boolean;
  deprecationDate?: string;
  sunsetDate?: string;
  migrationGuide?: string;
}

/**
 * Default API version configuration matching the backend
 */
export const DEFAULT_VERSION_CONFIG: VersionConfig = {
  defaultVersion: 'v1',
  deprecationWarning: true,
  supportedVersions: [
    {
      version: 'v1',
      isDefault: true,
      isDeprecated: false,
    },
  ],
};

/**
 * Current API version to use for requests
 * This should match the backend's default version
 */
export const CURRENT_API_VERSION = 'v1';

/**
 * Version detection strategies
 */
export enum VersionStrategy {
  HEADER = 'header',      // Accept: application/vnd.openagent.v2+json
  API_HEADER = 'api-header', // API-Version: v2
  QUERY = 'query',        // ?api-version=v2
  PATH = 'path',          // /api/v2/...
  DEFAULT = 'default',    // Use configured default
}

/**
 * Creates version headers for API requests
 */
export function createVersionHeaders(
  version: string = CURRENT_API_VERSION,
  strategy: VersionStrategy = VersionStrategy.HEADER,
): Record<string, string> {
  const headers: Record<string, string> = {};

  switch (strategy) {
    case VersionStrategy.HEADER:
      headers['Accept'] = `application/vnd.openagent.${version}+json`;
      break;
    case VersionStrategy.API_HEADER:
      headers['API-Version'] = version;
      break;
    // Query and path strategies are handled at the request level
  }

  return headers;
}

/**
 * Adds version parameter to URL query string
 */
export function addVersionQuery(url: string, version: string): string {
  const separator = url.includes('?') ? '&' : '?';
  return `${url}${separator}api-version=${version}`;
}

/**
 * Modifies URL path to include version
 */
export function addVersionPath(url: string, version: string): string {
  // Replace /api/ with /api/v{version}/
  return url.replace(/^\/api\//, `/api/${version}/`);
}

/**
 * Parses version information from response headers
 */
export function parseVersionResponse(headers: Record<string, string>): VersionResponse {
  const response: VersionResponse = {
    version: headers['api-version'] || headers['API-Version'] || CURRENT_API_VERSION,
  };

  if (headers['deprecation'] === 'true') {
    response.isDeprecated = true;
    
    if (headers['deprecation-date']) {
      response.deprecationDate = headers['deprecation-date'];
    }
    
    if (headers['sunset']) {
      response.sunsetDate = headers['sunset'];
    }
    
    // Extract migration guide URL from Link header
    const linkHeader = headers['link'];
    if (linkHeader) {
      const match = linkHeader.match(/<([^>]+)>;\s*rel="deprecation"/);
      if (match) {
        response.migrationGuide = match[1];
      }
    }
  }

  return response;
}

/**
 * Checks if a version is supported
 */
export function isVersionSupported(
  version: string,
  config: VersionConfig = DEFAULT_VERSION_CONFIG,
): boolean {
  return config.supportedVersions.some(v => v.version === version);
}

/**
 * Gets the default version from configuration
 */
export function getDefaultVersion(
  config: VersionConfig = DEFAULT_VERSION_CONFIG,
): string {
  return config.defaultVersion;
}

/**
 * Checks if a version is deprecated
 */
export function isVersionDeprecated(
  version: string,
  config: VersionConfig = DEFAULT_VERSION_CONFIG,
): boolean {
  const versionConfig = config.supportedVersions.find(v => v.version === version);
  return versionConfig?.isDeprecated || false;
}

/**
 * Gets deprecation information for a version
 */
export function getVersionDeprecationInfo(
  version: string,
  config: VersionConfig = DEFAULT_VERSION_CONFIG,
): { deprecatedAt?: Date; sunsetAt?: Date } | null {
  const versionConfig = config.supportedVersions.find(v => v.version === version);
  if (!versionConfig?.isDeprecated) {
    return null;
  }

  return {
    deprecatedAt: versionConfig.deprecatedAt ? new Date(versionConfig.deprecatedAt) : undefined,
    sunsetAt: versionConfig.sunsetAt ? new Date(versionConfig.sunsetAt) : undefined,
  };
}

/**
 * Feature detection based on API version
 * This helps determine what features are available in different versions
 */
export interface APIFeatures {
  batchEventIngestion: boolean;
  webhookSupport: boolean;
  advancedFiltering: boolean;
  realTimeUpdates: boolean;
}

export function getAPIFeatures(version: string): APIFeatures {
  // All features are available in v1 for now
  // This structure allows for feature flagging in future versions
  const baseFeatures: APIFeatures = {
    batchEventIngestion: true,
    webhookSupport: true,
    advancedFiltering: true,
    realTimeUpdates: true,
  };

  switch (version) {
    case 'v1':
      return baseFeatures;
    default:
      return baseFeatures;
  }
}

/**
 * Error types related to API versioning
 */
export class VersionError extends Error {
  constructor(
    message: string,
    public version: string,
    public supportedVersions?: string[],
  ) {
    super(message);
    this.name = 'VersionError';
  }
}

export class DeprecatedVersionError extends Error {
  constructor(
    message: string,
    public version: string,
    public deprecatedAt?: Date,
    public sunsetAt?: Date,
    public migrationGuide?: string,
  ) {
    super(message);
    this.name = 'DeprecatedVersionError';
  }
}

export class SunsetVersionError extends Error {
  constructor(
    message: string,
    public version: string,
    public sunsetAt: Date,
    public migrationGuide?: string,
  ) {
    super(message);
    this.name = 'SunsetVersionError';
  }
}

/**
 * Validates version compatibility and throws appropriate errors
 */
export function validateVersion(
  version: string,
  config: VersionConfig = DEFAULT_VERSION_CONFIG,
): void {
  if (!isVersionSupported(version, config)) {
    throw new VersionError(
      `API version ${version} is not supported`,
      version,
      config.supportedVersions.map(v => v.version),
    );
  }

  const versionConfig = config.supportedVersions.find(v => v.version === version);
  if (versionConfig?.isDeprecated) {
    const deprecationInfo = getVersionDeprecationInfo(version, config);
    
    // Check if version is sunset
    if (deprecationInfo?.sunsetAt && new Date() > deprecationInfo.sunsetAt) {
      throw new SunsetVersionError(
        `API version ${version} was sunset on ${deprecationInfo.sunsetAt.toDateString()}`,
        version,
        deprecationInfo.sunsetAt,
        '/docs/webui/api-versioning.md',
      );
    }
    
    // Version is deprecated but not sunset yet
    throw new DeprecatedVersionError(
      `API version ${version} is deprecated`,
      version,
      deprecationInfo?.deprecatedAt,
      deprecationInfo?.sunsetAt,
      '/docs/webui/api-versioning.md',
    );
  }
}