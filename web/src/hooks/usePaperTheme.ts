import { useEffect } from 'react';
import { useThemeStore } from '@/stores/themeStore';

// usePaperTheme declares this page "wants light mode by default" — but
// respects the user's explicit preference from the theme store. A page
// that never calls this (e.g. the editor) inherits the user's
// preference directly. Net result:
//
//   user pref = dark     → everything dark
//   user pref = light    → everything paper (including editor)
//   user pref = system   → OS decides
//
// Pages that historically forced paper now just defer to the store.
export function usePaperTheme() {
  const resolved = useThemeStore((s) => s.resolved);
  useEffect(() => {
    if (resolved === 'light') {
      document.body.classList.add('paper');
    } else {
      document.body.classList.remove('paper');
    }
  }, [resolved]);
}
