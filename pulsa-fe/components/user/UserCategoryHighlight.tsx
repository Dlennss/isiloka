'use client';

import { useEffect, useState } from 'react';

type LastVisitedCategory = {
  id: string;
  name: string;
  timestamp: number;
};

export function UserCategoryHighlight() {
  const [lastVisited, setLastVisited] = useState<LastVisitedCategory | null>(null);

  useEffect(() => {
    queueMicrotask(() => {
      const stored = localStorage.getItem('lastVisitedCategory');
      if (stored) {
        try {
          const parsed = JSON.parse(stored) as LastVisitedCategory;
          if (Date.now() - parsed.timestamp < 30 * 60 * 1000) {
            setLastVisited(parsed);
          } else {
            localStorage.removeItem('lastVisitedCategory');
          }
        } catch {
          localStorage.removeItem('lastVisitedCategory');
        }
      }
    });
  }, []);

  if (!lastVisited) return null;

  return (
    <div className="mb-3 rounded-lg bg-sky-50 p-3 text-sm text-sky-700">
      <span className="font-medium">Terakhir dilihat:</span> {lastVisited.name}
    </div>
  );
}
