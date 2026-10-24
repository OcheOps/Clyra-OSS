import type { Config } from 'tailwindcss';

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        // ── Marketing / auth: calm, editorial, premium ────────────
        // Warm-white paper surfaces. Used by landing, login, signup,
        // legal. The signed-in app stays on the dark `clyra-*` tokens
        // further down so the editor canvas works.
        paper: {
          DEFAULT: '#fbfaf7',  // page background — warm off-white
          soft:    '#f5f3ee',  // muted surface (section bands, muted cards)
          raised:  '#ffffff',  // raised cards / panels
          sunken:  '#efece5',  // sunken input background
        },

        // Deep charcoal / soft black for text on paper.
        charcoal: {
          DEFAULT: '#121316',  // strongest — headlines
          soft:    '#1c1e23',  // body
          muted:   '#5f6571',  // secondary text
          faint:   '#8d93a0',  // tertiary text, placeholders
        },

        // Warm-grey borders on paper surfaces.
        stone: {
          subtle:  '#ece9e1',  // hairline dividers
          default: '#dedad0',  // standard borders
          strong:  '#c7c2b5',  // emphasized / focused
        },

        // ── Signature accent: deep teal ────────────────────────────
        // Used sparingly — only on primary CTA, key stats, and
        // interactive focus. No gradients. Never on backgrounds.
        teal: {
          50:  '#e7f3f2',
          100: '#c6e3e1',
          200: '#9ccfcc',
          300: '#6db7b3',
          400: '#3f9a96',
          500: '#0f7d79',   // primary accent
          600: '#0a6562',
          700: '#084f4d',
          800: '#063a38',
          900: '#042725',
        },

        // ── Dark app surfaces (editor, dashboard) ──────────────────
        // Retained verbatim — the app canvas needs dark.
        brand: {
          50:  '#edfaff',
          100: '#d6f3ff',
          200: '#b5ebff',
          300: '#83dfff',
          400: '#48c9ff',
          500: '#1eaaff',
          600: '#068bff',
          700: '#0170e0',
          800: '#085bb6',
          900: '#0d4e8f',
          950: '#0e3157',
        },
        clyra: {
          'bg':         '#0a0d14',
          'base':       '#0f1219',
          'raised':     '#151a24',
          'elevated':   '#1b2130',
          'overlay':    '#212838',
        },
        line: {
          'subtle':     '#1e2433',
          'default':    '#252d3d',
          'strong':     '#2f3a4e',
          'accent':     '#3b4a66',
        },
        ink: {
          'primary':    '#e8ecf4',
          'secondary':  '#9ba4b8',
          'muted':      '#6b7488',
          'faint':      '#4a5268',
        },
        success:  { DEFAULT: '#34d399', muted: '#34d39920' },
        warning:  { DEFAULT: '#fbbf24', muted: '#fbbf2420' },
        danger:   { DEFAULT: '#f87171', muted: '#f8717120' },
        canvas: {
          bg:      '#0a0d14',
          surface: '#151a24',
          border:  '#252d3d',
          text:    '#e8ecf4',
          muted:   '#9ba4b8',
        },

        // ── v2 design system (locked palette B: saturated cyan + emerald
        // on slate). Side-by-side with the existing tokens above — every
        // legacy page keeps working unchanged because nothing in this
        // block touches an existing class name. New / restyled surfaces
        // opt in by using `bg-c2-bg`, `text-c2-text`, etc. As pages
        // migrate, the old tokens above can be deleted.
        c2: {
          'bg':            '#0F172A',  // root background (slate-900)
          'surface':       '#1E293B',  // cards, panels (slate-800)
          'elevated':      '#334155',  // popovers, dropdowns (slate-700)
          'overlay':       'rgba(15, 23, 42, 0.80)', // modal scrim

          'border':        '#334155',  // standard border (slate-700)
          'border-strong': '#475569',  // emphasized / focus (slate-600)
          'hairline':      '#1E293B',  // subtle dividers (slate-800)

          'text':          '#F8FAFC',  // primary text (slate-50)
          'text-muted':    '#94A3B8',  // secondary (slate-400)
          'text-faint':    '#64748B',  // tertiary (slate-500)

          'primary':       '#06B6D4',  // primary brand cyan (cyan-500)
          'primary-hover': '#0891B2',  // (cyan-600)
          'primary-soft':  'rgba(6, 182, 212, 0.15)',   // tinted bg
          'primary-ring':  'rgba(6, 182, 212, 0.30)',   // focus ring

          'accent':        '#10B981',  // accent emerald (emerald-500)
          'accent-hover':  '#059669',  // (emerald-600)
          'accent-soft':   'rgba(16, 185, 129, 0.15)',
          'accent-ring':   'rgba(16, 185, 129, 0.30)',

          // Status colors tuned for the slate palette
          'success':       '#34D399',  // emerald-400 (more readable than -500 on slate)
          'warning':       '#FBBF24',  // amber-400
          'danger':        '#F87171',  // red-400
          'info':          '#67E8F9',  // cyan-300 — distinct from primary
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
        // v2 type system. `font-display` = headlines / wordmarks in
        // Switzer with display-tracking. `font-brand` = general body on
        // v2 surfaces (same family, default tracking). Code stays mono.
        display: ['Switzer', 'Inter', 'system-ui', 'sans-serif'],
        brand:   ['Switzer', 'Inter', 'system-ui', 'sans-serif'],
      },
      fontSize: {
        'micro':  ['10px', { lineHeight: '14px' }],
        'tiny':   ['11px', { lineHeight: '16px' }],
        'petit':  ['12px', { lineHeight: '16px' }],
      },
      borderRadius: {
        'clyra':     '8px',
        'clyra-lg':  '12px',
        'clyra-xl':  '16px',
      },
      boxShadow: {
        'clyra-sm':   '0 1px 2px rgba(0, 0, 0, 0.3)',
        'clyra':      '0 2px 8px rgba(0, 0, 0, 0.4)',
        'clyra-lg':   '0 4px 16px rgba(0, 0, 0, 0.5)',
        'clyra-glow': '0 0 20px rgba(30, 170, 255, 0.15)',
        // Paper shadows — softer, warmer, for light-mode surfaces.
        'paper-sm':   '0 1px 2px rgba(18, 19, 22, 0.04), 0 0 0 1px rgba(18, 19, 22, 0.03)',
        'paper':      '0 2px 6px rgba(18, 19, 22, 0.05), 0 0 0 1px rgba(18, 19, 22, 0.04)',
        'paper-lg':   '0 12px 32px -8px rgba(18, 19, 22, 0.10), 0 0 0 1px rgba(18, 19, 22, 0.04)',

        // v2 shadow / glow scale. Glows are cyan-tinted to match the
        // palette B primary. Use sparingly — only on raised primary CTA
        // and the hero canvas. Everything else stays flat.
        'c2-sm':         '0 1px 2px rgba(0, 0, 0, 0.4)',
        'c2':            '0 4px 16px -4px rgba(0, 0, 0, 0.5)',
        'c2-lg':         '0 20px 60px -20px rgba(0, 0, 0, 0.6)',
        'c2-glow':       '0 0 32px -8px rgba(6, 182, 212, 0.35)',
        'c2-glow-soft':  '0 0 24px -10px rgba(6, 182, 212, 0.2)',
        'c2-accent-glow': '0 0 24px -8px rgba(16, 185, 129, 0.30)',
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      },
      letterSpacing: {
        'tight-ish': '-0.015em',
        'display':   '-0.025em',
      },
    },
  },
  plugins: [],
} satisfies Config;
