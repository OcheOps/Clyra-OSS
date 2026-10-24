import { useCallback, useState } from 'react';
import toast from 'react-hot-toast';
import { generateIaC } from '@/services/api/iac';
import { APIError } from '@/types/api';
import { useUpgradePromptStore } from '@/features/billing/upgradePromptStore';
import type { IaCCloud, IaCFormat, IaCGenerateResponse } from '@/types/iac';
import type { ProviderName } from '@/types/diagram';
import { useEditorStore } from '../store';

interface GenerateOptions {
  format: IaCFormat;
  cloud: IaCCloud;
  provider: ProviderName;
  apiKey?: string;
  explain?: boolean;
}

export function useIaC() {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<IaCGenerateResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const getDiagram = useEditorStore((s) => s.getDiagram);
  const mode = useEditorStore((s) => s.mode);

  const generate = useCallback(
    async ({ format, cloud, provider, apiKey, explain }: GenerateOptions) => {
      if (mode !== 'solution_architecture') {
        toast.error('IaC generation is only available in Solution mode');
        return null;
      }

      const diagram = getDiagram();
      if (diagram.nodes.length === 0) {
        toast.error('Add some nodes before generating infrastructure');
        return null;
      }

      setLoading(true);
      setError(null);
      try {
        const resp = await generateIaC({
          diagram,
          mode,
          cloud,
          format,
          provider,
          ...(apiKey ? { api_key: apiKey } : {}),
          ...(explain ? { explain: true } : {}),
        });
        setResult(resp);
        return resp;
      } catch (err) {
        if (err instanceof APIError && err.status === 402) {
          useUpgradePromptStore.getState().show('iac_export');
          setError('iac_export_gated');
          return null;
        }
        const msg = (err as Error).message ?? 'IaC generation failed';
        setError(msg);
        toast.error(msg);
        return null;
      } finally {
        setLoading(false);
      }
    },
    [mode, getDiagram],
  );

  const reset = useCallback(() => {
    setResult(null);
    setError(null);
  }, []);

  return { loading, result, error, generate, reset };
}
