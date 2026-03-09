import apiClient from './client';

export interface ReleaseAsset {
  name: string;
  browser_download_url: string;
  size: number;
}

export interface LatestRelease {
  tag_name: string;
  published_at: string;
  html_url: string;
  assets: ReleaseAsset[];
}

export async function getLatestRelease(): Promise<LatestRelease> {
  const { data } = await apiClient.get<LatestRelease>('/releases/latest');
  return data;
}
