import { HttpErrorResponse, HttpStatusCode } from '@angular/common/http';

export interface ApiFailure {
  fieldErrors: unknown;
  message: string;
}

export function isClientError(status: number): boolean {
  return (
    status >= HttpStatusCode.BadRequest && status < HttpStatusCode.InternalServerError
  );
}

export function readApiFailure(
  response: HttpErrorResponse,
  fallback: string
): ApiFailure {
  if (!isClientError(response.status)) {
    return { fieldErrors: null, message: fallback };
  }

  const message = response.error?.message;

  return {
    fieldErrors: response.error?.errors ?? null,
    message: typeof message === 'string' && message ? message : fallback
  };
}
