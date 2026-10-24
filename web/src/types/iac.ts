import type { Diagram, ProviderName, ArchitectureMode } from './diagram';

export type IaCFormat = 'terraform' | 'cloudformation';
export type IaCCloud = 'aws' | 'gcp' | 'azure';

export interface IaCGenerateRequest {
  diagram: Diagram;
  mode: ArchitectureMode;
  cloud: IaCCloud;
  format: IaCFormat;
  provider: ProviderName;
  api_key?: string;
  explain?: boolean;
}

export interface IaCGenerateResponse {
  code: string;
  format: IaCFormat;
  language: string;
  warnings?: string[];
  explanation?: string;
}
