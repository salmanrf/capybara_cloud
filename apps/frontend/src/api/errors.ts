/**
 * @params status: the HTTP status, or 0 when the request never got a response (network down, proxy not running); message: the backend's message; context: extra error details
 * @return an ApiError
 * The only error the api module throws.
 */
export class ApiError extends Error {
  readonly status: number;
  readonly context: Record<string, unknown>;

  constructor(status: number, message: string, context: Record<string, unknown> = {}) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.context = context;
  }
}
