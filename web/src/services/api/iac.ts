import { api } from './client';
import type { IaCGenerateRequest, IaCGenerateResponse } from '@/types/iac';

export function generateIaC(req: IaCGenerateRequest): Promise<IaCGenerateResponse> {
  return api<IaCGenerateResponse>('/iac/generate', {
    method: 'POST',
    body: req,
  });
}
