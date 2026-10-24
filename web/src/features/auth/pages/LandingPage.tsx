import { Link } from 'react-router-dom';
import { PricingSection } from '../components/PricingSection';
import { HeroDemo } from '../components/HeroDemo';

/**
 * Landing — v2.
 *
 * Aesthetic: refined-instrument. Dark slate substrate (palette B locked
 * in design-tokens.md), Switzer for type, JetBrains Mono carries the
 * eyebrow / data voice. Restraint everywhere except the hero canvas,
 * which is the one orchestrated motion moment per the brand brief.
 *
 * Layout breaks the grid in two places: hero is asymmetric 5/7 with the
 * canvas spilling into the right margin, and the "how it works" cards
 * stagger their vertical alignment so the row reads as a path, not a
 * static row of tiles.
 *
 * Voice tightens from "Think in diagrams. Ship in infrastructure." to
 * "Architecture, as it should be." Sub-copy carries the concrete value.
 *
 * usePaperTheme is intentionally NOT called here — palette B is dark
 * uniform across every marketing surface per the round-3 lock.
 */
export function LandingPage() {
  return (
    <div className="relative min-h-screen overflow-x-hidden bg-c2-bg font-brand text-c2-text antialiased">
      {/* Atmospheric background: a calm cyan haze at the top, a dot grid
          everywhere. The grid is the engineering substrate metaphor —
          this whole tool sits on a coordinate plane. */}
      <div
        aria-hidden
        className="pointer-events-none fixed inset-0"
        style={{
          backgroundImage:
            'radial-gradient(circle at 1px 1px, rgba(148,163,184,0.07) 1px, transparent 0)',
          backgroundSize: '24px 24px',
        }}
      />
      <div
        aria-hidden
        className="pointer-events-none fixed -left-1/4 -top-1/4 h-[700px] w-[1200px] opacity-50 blur-[120px]"
        style={{
          background:
            'radial-gradient(ellipse at center, rgba(6,182,212,0.18) 0%, transparent 60%)',
        }}
      />
      <div
        aria-hidden
        className="pointer-events-none fixed -right-1/4 top-1/3 h-[500px] w-[800px] opacity-30 blur-[120px]"
        style={{
          background:
            'radial-gradient(ellipse at center, rgba(16,185,129,0.14) 0%, transparent 60%)',
        }}
      />

      <div className="relative">
        <Nav />
        <Hero />
        <HowItWorks />
        <ForEngineers />
        <PricingSection />
        <ClosingCTA />
        <Footer />
      </div>
    </div>
  );
}

// ────────────────────────────────────────────────────────────────────
// Nav

function Nav() {
  return (
    <nav className="relative z-10 mx-auto flex max-w-6xl items-center justify-between px-6 py-6">
      <Wordmark />
      <div className="flex items-center gap-1">
        <NavLink href="#how">How it works</NavLink>
        <NavLink href="#pricing">Pricing</NavLink>
        <span className="mx-2 h-4 w-px bg-c2-border" aria-hidden />
        <Link
          to="/login"
          className="rounded-full px-3 py-1.5 text-tiny text-c2-text-muted transition-colors hover:text-c2-text"
        >
          Sign in
        </Link>
        <Link
          to="/signup"
          className="ml-1 rounded-full bg-c2-text px-3 py-1.5 text-tiny font-medium text-c2-bg transition-colors hover:bg-c2-text-muted"
        >
          Start free
        </Link>
      </div>
    </nav>
  );
}

function NavLink({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <a
      href={href}
      className="rounded-full px-3 py-1.5 text-tiny text-c2-text-muted transition-colors hover:text-c2-text"
    >
      {children}
    </a>
  );
}

// ────────────────────────────────────────────────────────────────────
// Hero — asymmetric 5/7 with canvas in the wider right column.

function Hero() {
  return (
    <section className="relative mx-auto max-w-6xl px-6 pt-16 pb-28 md:pt-24">
      <div className="grid items-center gap-14 md:grid-cols-12 md:gap-10">
        <div className="md:col-span-5">
          <Eyebrow>For engineers shipping on AWS</Eyebrow>
          <h1 className="mt-6 font-display text-[44px] font-semibold leading-[1.02] tracking-[-0.035em] text-c2-text md:text-[58px]">
            From prompt to
            <br />
            <span className="bg-gradient-to-r from-c2-primary via-c2-info to-c2-accent bg-clip-text text-transparent">
              Terraform in a minute.
            </span>
          </h1>
          <p className="mt-6 max-w-md text-[15px] leading-relaxed text-c2-text-muted">
            Describe your system in plain English. Get an editable
            architecture diagram <em className="not-italic text-c2-text">and</em>{' '}
            deployable Terraform &mdash; VPC, IAM, security groups, all of it.
            Skip the boilerplate; ship the design.
          </p>

          <div className="mt-9 flex flex-wrap items-center gap-3">
            <Link
              to="/signup"
              className="group relative inline-flex items-center gap-2 rounded-full bg-c2-primary px-5 py-2.5 text-sm font-medium text-c2-bg shadow-c2-glow transition-all hover:bg-c2-primary-hover hover:shadow-[0_0_40px_-6px_rgba(6,182,212,0.55)]"
            >
              Start free &mdash; no card
              <span className="transition-transform group-hover:translate-x-0.5">→</span>
            </Link>
            <Link
              // Signup → editor with the reverse-ingest modal auto-opened.
              // encoded so ?redirect= survives GuestGuard's URL-decode
              // and lands on /editor?reverse=1 whether the user signs
              // up fresh or is already logged in.
              to="/signup?redirect=%2Feditor%3Freverse%3D1"
              className="inline-flex items-center gap-2 rounded-full border border-c2-border bg-c2-surface/40 px-5 py-2.5 text-sm font-medium text-c2-text transition-colors hover:border-c2-border-strong hover:bg-c2-surface"
              title="Paste your existing Terraform and watch Clyra map it in seconds."
            >
              Import your Terraform
            </Link>
          </div>

          <div className="mt-5 flex flex-wrap items-center gap-x-3 gap-y-1 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            <span>60-second generations</span>
            <span>·</span>
            <span>terraform + cloudformation</span>
            <span>·</span>
            <span>free forever tier</span>
          </div>
        </div>

        {/* Canvas stays inside its column — the earlier negative-margin
            'bleed' caused horizontal overflow on narrow widths even with
            overflow-x-hidden on the root, which clipped pricing cards
            further down the page. The atmospheric glow already gives
            the hero plenty of presence; the bleed wasn't carrying its
            weight. */}
        <div className="md:col-span-7">
          <HeroDemo />
        </div>
      </div>
    </section>
  );
}

// ────────────────────────────────────────────────────────────────────
// How it works — three numbered cards. Vertical-stagger breaks the grid.

function HowItWorks() {
  return (
    <section
      id="how"
      className="relative border-t border-c2-hairline bg-c2-bg/60 backdrop-blur-sm"
    >
      <div className="mx-auto max-w-6xl px-6 py-24">
        <div className="max-w-2xl">
          <Eyebrow>How it works</Eyebrow>
          <h2 className="mt-5 font-display text-3xl font-semibold tracking-[-0.025em] text-c2-text md:text-[40px] md:leading-[1.1]">
            From idea to infrastructure
            <br />
            <span className="text-c2-text-muted">in three steps.</span>
          </h2>
        </div>

        <div className="mt-16 grid gap-6 md:grid-cols-3">
          <Step
            number="01"
            title="Describe"
            body="Say what you're building in plain English. Three-tier web app on AWS. Event pipeline. Static site with auth. Clyra interprets intent, not buzzwords."
            offsetClass="md:translate-y-0"
          />
          <Step
            number="02"
            title="Iterate"
            body="Review the generated diagram on a real canvas. Add nodes, re-wire edges, rename components. Follow-up prompts refine what's already there."
            offsetClass="md:translate-y-8"
          />
          <Step
            number="03"
            title="Ship code"
            body="Export starter Terraform or CloudFormation from the same diagram. A scaffold a senior engineer can finish — not a one-click deploy pretending to be one."
            offsetClass="md:translate-y-16"
          />
        </div>
      </div>
    </section>
  );
}

function Step({
  number,
  title,
  body,
  offsetClass,
}: {
  number: string;
  title: string;
  body: string;
  offsetClass: string;
}) {
  return (
    <article
      className={`group relative overflow-hidden rounded-2xl border border-c2-border bg-c2-surface/70 p-7 backdrop-blur-sm transition-all duration-300 hover:border-c2-border-strong hover:bg-c2-surface ${offsetClass}`}
    >
      <div className="flex items-center gap-3">
        <span className="flex h-9 w-9 items-center justify-center rounded-lg border border-c2-primary/30 bg-c2-primary-soft font-mono text-[11px] font-semibold text-c2-primary">
          {number}
        </span>
        <span className="h-px flex-1 bg-c2-hairline" />
      </div>
      <h3 className="mt-6 font-display text-lg font-semibold tracking-[-0.015em] text-c2-text">
        {title}
      </h3>
      <p className="mt-3 text-sm leading-relaxed text-c2-text-muted">{body}</p>

      {/* Subtle hover halo */}
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 opacity-0 transition-opacity duration-500 group-hover:opacity-100"
        style={{
          background:
            'radial-gradient(circle at 20% 0%, rgba(6,182,212,0.08) 0%, transparent 50%)',
        }}
      />
    </article>
  );
}

// ────────────────────────────────────────────────────────────────────
// For engineers — fact grid, denser composition with a stronger
// statement panel on the left.

function ForEngineers() {
  return (
    <section className="relative">
      <div className="mx-auto max-w-6xl px-6 py-24">
        <div className="grid gap-12 md:grid-cols-12 md:items-start md:gap-10">
          <div className="md:col-span-5 md:sticky md:top-24">
            <Eyebrow>Built for engineers</Eyebrow>
            <h2 className="mt-5 font-display text-3xl font-semibold leading-[1.1] tracking-[-0.025em] text-c2-text md:text-[40px]">
              A workbench —
              <br />
              <span className="text-c2-text-muted">not a toy.</span>
            </h2>
            <p className="mt-6 text-[15px] leading-relaxed text-c2-text-muted">
              Clyra mirrors how engineers actually design systems: reason about
              the shape, map it visually, then commit the infrastructure. Your
              diagram is plain JSON. Your IaC is code you own. Nothing
              proprietary in between.
            </p>
          </div>
          <div className="grid grid-cols-1 gap-4 md:col-span-7 md:grid-cols-2">
            <Fact
              title="AWS-native"
              body="Structured Terraform and CloudFormation output, golden-tested."
              tag="output"
            />
            <Fact
              title="IaC-ready"
              body="Starter scaffolds pass terraform validate and cfn-lint. One free export; unlimited on Pro."
              tag="lint"
            />
            <Fact
              title="BYO models"
              body="OpenAI, Anthropic, Gemini — your keys, your rate limits."
              tag="byo"
            />
            <Fact
              title="JSON-native"
              body="Your diagram is data. Version it. Diff it. Commit it."
              tag="data"
            />
            <Fact
              title="Long-form prompts"
              body="Free plans get 2,000 characters per prompt — plenty for everyday architectures. Pro and BYO unlock 8,000 for richly specified, multi-zone systems."
              tag="prompts"
              wide
            />
          </div>
        </div>
      </div>
    </section>
  );
}

function Fact({
  title,
  body,
  tag,
  wide,
}: {
  title: string;
  body: string;
  tag: string;
  wide?: boolean;
}) {
  return (
    <div
      className={`group rounded-xl border border-c2-border bg-c2-surface/50 p-5 transition-colors duration-200 hover:border-c2-primary/50 hover:bg-c2-surface ${
        wide ? 'md:col-span-2' : ''
      }`}
    >
      <div className="flex items-baseline justify-between gap-3">
        <h3 className="font-display text-[15px] font-semibold tracking-[-0.005em] text-c2-text">
          {title}
        </h3>
        <span className="font-mono text-[9px] uppercase tracking-[0.18em] text-c2-text-faint transition-colors group-hover:text-c2-primary">
          {tag}
        </span>
      </div>
      <p className="mt-2 text-[13px] leading-relaxed text-c2-text-muted">{body}</p>
    </div>
  );
}

// ────────────────────────────────────────────────────────────────────
// Closing CTA — big and quiet.

function ClosingCTA() {
  return (
    <section className="relative border-t border-c2-hairline">
      <div className="mx-auto max-w-3xl px-6 py-32 text-center">
        <Eyebrow center>Get started</Eyebrow>
        <h2 className="mt-6 font-display text-[42px] font-semibold leading-[1.05] tracking-[-0.03em] text-c2-text md:text-[56px]">
          A tool for thinking
          <br />
          <span className="bg-gradient-to-r from-c2-primary to-c2-accent bg-clip-text text-transparent">
            clearly.
          </span>
        </h2>
        <p className="mx-auto mt-6 max-w-md text-[15px] leading-relaxed text-c2-text-muted">
          Ten diagrams per account, free. Upgrade to Pro for 200, or bring your
          own AI keys for 1,000.
        </p>
        <div className="mt-10 flex items-center justify-center gap-3">
          <Link
            to="/signup"
            className="group inline-flex items-center gap-2 rounded-full bg-c2-primary px-6 py-3 text-sm font-medium text-c2-bg shadow-c2-glow transition-all hover:bg-c2-primary-hover hover:shadow-[0_0_40px_-6px_rgba(6,182,212,0.55)]"
          >
            Create your account
            <span className="transition-transform group-hover:translate-x-0.5">→</span>
          </Link>
          <Link
            to="/login"
            className="inline-flex items-center gap-2 rounded-full border border-c2-border bg-c2-surface/40 px-6 py-3 text-sm font-medium text-c2-text transition-colors hover:border-c2-border-strong hover:bg-c2-surface"
          >
            Sign in
          </Link>
        </div>
      </div>
    </section>
  );
}

// ────────────────────────────────────────────────────────────────────
// Footer — minimal, mono.

function Footer() {
  return (
    <footer className="relative border-t border-c2-hairline">
      <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-4 px-6 py-10 md:flex-row">
        <div className="flex items-center gap-3">
          <Wordmark small />
          <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            © {new Date().getFullYear()} · architecture → infrastructure
          </span>
        </div>
        <div className="flex items-center gap-5 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-muted">
          <Link to="/terms" className="transition-colors hover:text-c2-text">Terms</Link>
          <Link to="/privacy" className="transition-colors hover:text-c2-text">Privacy</Link>
          <Link to="/support" className="transition-colors hover:text-c2-text">Support</Link>
        </div>
      </div>
    </footer>
  );
}

// ────────────────────────────────────────────────────────────────────
// Shared primitives

function Eyebrow({
  children,
  center,
}: {
  children: React.ReactNode;
  center?: boolean;
}) {
  return (
    <div className={`flex items-center gap-2 ${center ? 'justify-center' : ''}`}>
      <span className="inline-block h-1.5 w-1.5 rounded-full bg-c2-primary shadow-[0_0_8px_rgba(6,182,212,0.6)]" />
      <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-c2-text-faint">
        {children}
      </span>
    </div>
  );
}

function Wordmark({ small = false }: { small?: boolean }) {
  const size = small ? 'text-base' : 'text-lg';
  return (
    <Link
      to="/"
      className={`${size} font-display font-semibold tracking-[-0.015em] text-c2-text`}
    >
      clyra<span className="text-c2-primary">.</span>
    </Link>
  );
}
