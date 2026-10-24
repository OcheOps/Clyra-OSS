export interface APIResponse<T> {
  data?: T;
  error?: string;
  details?: string[];
}

export class APIError extends Error {
  constructor(
    public status: number,
    message: string,
    public details?: string[],
  ) {
    super(message);
    this.name = 'APIError';
  }
}
