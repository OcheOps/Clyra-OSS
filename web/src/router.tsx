import { lazy } from 'react';
import { createBrowserRouter } from 'react-router-dom';
import { AuthGuard } from '@/features/auth/components/AuthGuard';
import { GuestGuard } from '@/features/auth/components/GuestGuard';

const LandingPage = lazy(() => import('@/features/auth/pages/LandingPage').then((m) => ({ default: m.LandingPage })));
const LoginPage = lazy(() => import('@/features/auth/pages/LoginPage').then((m) => ({ default: m.LoginPage })));
const SignupPage = lazy(() => import('@/features/auth/pages/SignupPage').then((m) => ({ default: m.SignupPage })));
const ForgotPasswordPage = lazy(() => import('@/features/auth/pages/ForgotPasswordPage').then((m) => ({ default: m.ForgotPasswordPage })));
const ResetPasswordPage = lazy(() => import('@/features/auth/pages/ResetPasswordPage').then((m) => ({ default: m.ResetPasswordPage })));
const TermsPage = lazy(() => import('@/features/legal/pages/TermsPage').then((m) => ({ default: m.TermsPage })));
const PrivacyPage = lazy(() => import('@/features/legal/pages/PrivacyPage').then((m) => ({ default: m.PrivacyPage })));
const BillingPage = lazy(() => import('@/features/billing/pages/BillingPage').then((m) => ({ default: m.BillingPage })));
const BillingReturnPage = lazy(() => import('@/features/billing/pages/BillingReturnPage').then((m) => ({ default: m.BillingReturnPage })));
const DashboardPage = lazy(() => import('@/features/dashboard/pages/DashboardPage').then((m) => ({ default: m.DashboardPage })));
const EditorPage = lazy(() => import('@/features/editor/pages/EditorPage').then((m) => ({ default: m.EditorPage })));
const SettingsPage = lazy(() => import('@/features/settings/pages/SettingsPage').then((m) => ({ default: m.SettingsPage })));
const AdminPage = lazy(() => import('@/features/admin/pages/AdminPage').then((m) => ({ default: m.AdminPage })));
const ProjectsPage = lazy(() => import('@/features/projects/pages/ProjectsPage').then((m) => ({ default: m.ProjectsPage })));
const PublicSharePage = lazy(() => import('@/features/share/pages/PublicSharePage').then((m) => ({ default: m.PublicSharePage })));
const SupportPage = lazy(() => import('@/features/support/pages/SupportPage').then((m) => ({ default: m.SupportPage })));

export const router = createBrowserRouter([
  {
    path: '/',
    element: (
      <GuestGuard>
        <LandingPage />
      </GuestGuard>
    ),
  },
  {
    path: '/login',
    element: (
      <GuestGuard>
        <LoginPage />
      </GuestGuard>
    ),
  },
  {
    path: '/signup',
    element: (
      <GuestGuard>
        <SignupPage />
      </GuestGuard>
    ),
  },
  {
    path: '/forgot-password',
    element: (
      <GuestGuard>
        <ForgotPasswordPage />
      </GuestGuard>
    ),
  },
  {
    path: '/reset-password',
    // No guard — an authenticated user who follows a reset link from email
    // should still land here (they might be fixing a different account).
    element: <ResetPasswordPage />,
  },
  {
    path: '/terms',
    element: <TermsPage />,
  },
  {
    path: '/privacy',
    element: <PrivacyPage />,
  },
  {
    path: '/billing',
    element: (
      <AuthGuard>
        <BillingPage />
      </AuthGuard>
    ),
  },
  {
    path: '/billing/return',
    element: (
      <AuthGuard>
        <BillingReturnPage />
      </AuthGuard>
    ),
  },
  {
    path: '/dashboard',
    element: (
      <AuthGuard>
        <DashboardPage />
      </AuthGuard>
    ),
  },
  {
    path: '/editor/:projectId?',
    element: (
      <AuthGuard>
        <EditorPage />
      </AuthGuard>
    ),
  },
  {
    path: '/settings',
    element: (
      <AuthGuard>
        <SettingsPage />
      </AuthGuard>
    ),
  },
  {
    path: '/projects',
    element: (
      <AuthGuard>
        <ProjectsPage />
      </AuthGuard>
    ),
  },
  {
    path: '/admin',
    element: (
      <AuthGuard>
        <AdminPage />
      </AuthGuard>
    ),
  },
  {
    // Public share viewer — no auth.
    path: '/share/:token',
    element: <PublicSharePage />,
  },
  {
    // Public support / feedback form. Pre-fills user info when signed in.
    path: '/support',
    element: <SupportPage />,
  },
]);
