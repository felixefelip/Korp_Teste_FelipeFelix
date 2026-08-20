import { HttpStatusCode } from '@angular/common/http';

export function isClientError(status: number): boolean {
  return (
    status >= HttpStatusCode.BadRequest && status < HttpStatusCode.InternalServerError
  );
}
