import { api } from './client';

export type RegistrationMode = 'invite' | 'open';

export interface PublicConfig {
  registration_mode: RegistrationMode;
  google_oauth_enabled: boolean;
  github_oauth_enabled: boolean;
}

export async function getPublicConfig(): Promise<PublicConfig> {
  return api<PublicConfig>('/config');
}
