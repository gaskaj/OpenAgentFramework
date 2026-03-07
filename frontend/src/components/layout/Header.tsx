import { useCallback, useEffect, useState } from 'react';
import * as DropdownMenu from '@radix-ui/react-dropdown-menu';
import { ChevronDown, LogOut, User as UserIcon, Building2 } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { listOrgs } from '@/api/orgs';
import type { Organization } from '@/types';

export function Header() {
  const { user, currentOrg, logout, setCurrentOrg } = useAuth();
  const [orgs, setOrgs] = useState<Organization[]>([]);

  useEffect(() => {
    listOrgs()
      .then(setOrgs)
      .catch(() => {});
  }, []);

  const switchOrg = useCallback(
    (org: Organization) => {
      setCurrentOrg(org);
    },
    [setCurrentOrg],
  );

  return (
    <header className="flex h-16 items-center justify-between border-b border-zinc-700 bg-zinc-900 px-6">
      {/* Org selector */}
      <DropdownMenu.Root>
        <DropdownMenu.Trigger asChild>
          <button className="flex items-center gap-2 rounded-lg border border-zinc-700 px-3 py-1.5 text-sm text-zinc-300 transition-colors hover:bg-zinc-800">
            <Building2 className="h-4 w-4 text-zinc-500" />
            <span>{currentOrg?.name ?? 'Select organization'}</span>
            <ChevronDown className="h-4 w-4 text-zinc-500" />
          </button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Portal>
          <DropdownMenu.Content
            className="min-w-[200px] rounded-lg border border-zinc-700 bg-zinc-800 p-1 shadow-xl"
            sideOffset={8}
          >
            {orgs.map((org) => (
              <DropdownMenu.Item
                key={org.id}
                onSelect={() => switchOrg(org)}
                className="cursor-pointer rounded-md px-3 py-2 text-sm text-zinc-300 outline-none hover:bg-zinc-700 focus:bg-zinc-700"
              >
                {org.name}
              </DropdownMenu.Item>
            ))}
          </DropdownMenu.Content>
        </DropdownMenu.Portal>
      </DropdownMenu.Root>

      {/* User menu */}
      <DropdownMenu.Root>
        <DropdownMenu.Trigger asChild>
          <button className="flex items-center gap-2 rounded-lg px-2 py-1.5 text-sm text-zinc-300 transition-colors hover:bg-zinc-800">
            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 text-xs font-medium text-white">
              {user?.display_name?.charAt(0)?.toUpperCase() ?? 'U'}
            </div>
            <span className="hidden sm:inline">{user?.display_name ?? user?.email}</span>
            <ChevronDown className="h-4 w-4 text-zinc-500" />
          </button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Portal>
          <DropdownMenu.Content
            className="min-w-[180px] rounded-lg border border-zinc-700 bg-zinc-800 p-1 shadow-xl"
            sideOffset={8}
            align="end"
          >
            <DropdownMenu.Item className="cursor-pointer rounded-md px-3 py-2 text-sm text-zinc-300 outline-none hover:bg-zinc-700 focus:bg-zinc-700">
              <div className="flex items-center gap-2">
                <UserIcon className="h-4 w-4" />
                Profile
              </div>
            </DropdownMenu.Item>
            <DropdownMenu.Separator className="my-1 h-px bg-zinc-700" />
            <DropdownMenu.Item
              onSelect={logout}
              className="cursor-pointer rounded-md px-3 py-2 text-sm text-red-400 outline-none hover:bg-zinc-700 focus:bg-zinc-700"
            >
              <div className="flex items-center gap-2">
                <LogOut className="h-4 w-4" />
                Log out
              </div>
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Portal>
      </DropdownMenu.Root>
    </header>
  );
}
