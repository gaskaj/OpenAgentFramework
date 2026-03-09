import { useState, useEffect, useMemo } from 'react';
import { Download, Copy, Check, Terminal, ExternalLink } from 'lucide-react';
import { getLatestRelease, type LatestRelease } from '@/api/releases';

interface Props {
  apiKey: string;
  controlPlaneUrl: string;
}

type Platform = 'darwin-arm64' | 'darwin-amd64' | 'linux-amd64' | 'linux-arm64' | 'windows-amd64';

const PLATFORM_INFO: Record<Platform, { label: string; shortLabel: string; icon: string }> = {
  'darwin-arm64': { label: 'macOS (Apple Silicon)', shortLabel: 'macOS ARM', icon: '' },
  'darwin-amd64': { label: 'macOS (Intel)', shortLabel: 'macOS Intel', icon: '' },
  'linux-amd64':  { label: 'Linux (x86_64)', shortLabel: 'Linux x64', icon: '' },
  'linux-arm64':  { label: 'Linux (ARM64)', shortLabel: 'Linux ARM', icon: '' },
  'windows-amd64': { label: 'Windows (x86_64)', shortLabel: 'Windows', icon: '' },
};

const PLATFORM_ORDER: Platform[] = [
  'darwin-arm64',
  'darwin-amd64',
  'linux-amd64',
  'linux-arm64',
  'windows-amd64',
];

function detectPlatform(): Platform {
  const ua = navigator.userAgent.toLowerCase();
  if (ua.includes('win')) return 'windows-amd64';
  if (ua.includes('mac')) {
    // Rough heuristic: recent macOS is likely Apple Silicon
    return 'darwin-arm64';
  }
  return 'linux-amd64';
}

function assetForPlatform(assets: LatestRelease['assets'], platform: Platform): LatestRelease['assets'][0] | undefined {
  // Asset names are like: agentctl-linux-amd64, agentctl-darwin-arm64, agentctl-windows-amd64.exe
  const [os, arch] = platform.split('-');
  return assets.find((a) => a.name.includes(`-${os}-${arch}`));
}

function formatSize(bytes: number): string {
  if (bytes === 0) return '';
  const mb = bytes / (1024 * 1024);
  return `${mb.toFixed(1)} MB`;
}

export function DownloadInstructions({ apiKey, controlPlaneUrl }: Props) {
  const [release, setRelease] = useState<LatestRelease | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedPlatform, setSelectedPlatform] = useState<Platform>(detectPlatform);
  const [copiedScript, setCopiedScript] = useState(false);

  useEffect(() => {
    getLatestRelease()
      .then(setRelease)
      .catch(() => setRelease(null))
      .finally(() => setLoading(false));
  }, []);

  const asset = useMemo(
    () => (release?.assets ? assetForPlatform(release.assets, selectedPlatform) : undefined),
    [release, selectedPlatform],
  );

  const isWindows = selectedPlatform === 'windows-amd64';
  const binaryName = isWindows ? 'agentctl.exe' : 'agentctl';

  const setupScript = useMemo(() => {
    if (!asset) return '';
    const downloadUrl = asset.browser_download_url;

    if (isWindows) {
      return `# Download agentctl
Invoke-WebRequest -Uri "${downloadUrl}" -OutFile "agentctl.exe"

# Write config
@"
controlplane:
  enabled: true
  url: "${controlPlaneUrl}"
  api_key: "${apiKey}"
  config_mode: "remote"
  config_poll_interval: "30s"
"@ | Set-Content -Path "config.yaml"

# Start the agent
.\\agentctl.exe start --config config.yaml`;
    }

    return `# Download agentctl
curl -Lo ${binaryName} "${downloadUrl}"
chmod +x ${binaryName}

# Write config
cat > config.yaml <<'EOF'
controlplane:
  enabled: true
  url: "${controlPlaneUrl}"
  api_key: "${apiKey}"
  config_mode: "remote"
  config_poll_interval: "30s"
EOF

# Start the agent
./${binaryName} start --config config.yaml`;
  }, [asset, apiKey, controlPlaneUrl, binaryName, isWindows]);

  const copyScript = async () => {
    await navigator.clipboard.writeText(setupScript);
    setCopiedScript(true);
    setTimeout(() => setCopiedScript(false), 2000);
  };

  if (loading) {
    return (
      <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-6">
        <div className="flex items-center gap-3 text-zinc-400">
          <Download className="h-5 w-5 animate-pulse" />
          <span className="text-sm">Loading release information...</span>
        </div>
      </div>
    );
  }

  // No release published yet
  if (!release || !release.tag_name) {
    return (
      <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-6">
        <div className="flex items-center gap-3">
          <Terminal className="h-5 w-5 text-zinc-400" />
          <div>
            <h2 className="text-lg font-semibold text-zinc-100">Install agentctl</h2>
            <p className="text-sm text-zinc-400">Build from source</p>
          </div>
        </div>
        <pre className="mt-4 overflow-x-auto rounded-lg bg-zinc-900 p-4 text-xs text-zinc-400">
{`git clone https://github.com/gaskaj/OpenAgentFramework.git
cd OpenAgentFramework
make build
./bin/agentctl start --config config.yaml`}
        </pre>
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Download className="h-5 w-5 text-blue-400" />
          <div>
            <h2 className="text-lg font-semibold text-zinc-100">Download & Run</h2>
            <p className="text-sm text-zinc-400">
              agentctl{' '}
              <a
                href={release.html_url}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 text-blue-400 hover:text-blue-300"
              >
                {release.tag_name}
                <ExternalLink className="h-3 w-3" />
              </a>
            </p>
          </div>
        </div>
      </div>

      {/* Platform tabs */}
      <div className="mt-4 flex flex-wrap gap-1.5">
        {PLATFORM_ORDER.map((p) => {
          const info = PLATFORM_INFO[p];
          const platformAsset = assetForPlatform(release.assets, p);
          if (!platformAsset) return null;
          return (
            <button
              key={p}
              onClick={() => setSelectedPlatform(p)}
              className={`rounded-md px-3 py-1.5 text-xs font-medium transition-colors ${
                selectedPlatform === p
                  ? 'bg-blue-600 text-white'
                  : 'bg-zinc-900 text-zinc-400 hover:bg-zinc-700 hover:text-zinc-200'
              }`}
            >
              {info.shortLabel}
            </button>
          );
        })}
      </div>

      {/* Download link */}
      {asset && (
        <div className="mt-3">
          <a
            href={asset.browser_download_url}
            className="inline-flex items-center gap-2 rounded-lg border border-blue-500/30 bg-blue-500/10 px-4 py-2 text-sm font-medium text-blue-300 transition-colors hover:bg-blue-500/20"
          >
            <Download className="h-4 w-4" />
            Download {asset.name}
            {asset.size > 0 && (
              <span className="text-xs text-blue-400/70">({formatSize(asset.size)})</span>
            )}
          </a>
        </div>
      )}

      {/* Setup script */}
      {setupScript && (
        <div className="mt-4">
          <div className="flex items-center justify-between">
            <label className="text-xs font-medium text-zinc-400">
              Quick setup ({PLATFORM_INFO[selectedPlatform].label})
            </label>
            <button
              onClick={copyScript}
              className="flex items-center gap-1 text-xs text-zinc-500 transition-colors hover:text-zinc-300"
            >
              {copiedScript ? (
                <>
                  <Check className="h-3 w-3" /> Copied
                </>
              ) : (
                <>
                  <Copy className="h-3 w-3" /> Copy
                </>
              )}
            </button>
          </div>
          <pre className="mt-1.5 overflow-x-auto rounded-lg bg-zinc-900 p-4 text-xs text-zinc-400">
            {setupScript}
          </pre>
        </div>
      )}
    </div>
  );
}
