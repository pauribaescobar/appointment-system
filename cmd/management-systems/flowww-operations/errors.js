class AppError extends Error {
  constructor(kind, code, message, opts = {}) {
    super(message);
    this.kind = kind;           // TRANSIENT, UNAUTHORIZED...
    this.code = code;           // FLOWWW_TIMEOUT, FLOWWW_LOGIN_FAILED...
    this.details = opts.details;
  }
}

function mapError(err) {
  // Puppeteer timeout suele ser name = 'TimeoutError'
  if (err?.name === 'TimeoutError') {
    return new AppError('TRANSIENT', 'FLOWWW_TIMEOUT', err.message);
  }

  // errores típicos de red
  const msg = String(err?.message || '');
  if (msg.includes('net::') || msg.includes('ECONNRESET') || msg.includes('ETIMEDOUT')) {
    return new AppError('TRANSIENT', 'NETWORK_ERROR', msg);
  }

  // login / auth (si detectas por texto o por validación tuya)
  if (msg.toLowerCase().includes('login')) {
    return new AppError('UNAUTHORIZED', 'FLOWWW_LOGIN_FAILED', msg);
  }

  // input
  if (err?.code === 'INVALID_CENTER') {
    return new AppError('INVALID_INPUT', 'INVALID_CENTER', msg);
  }

  return new AppError('INTERNAL', 'UNEXPECTED', msg || 'Unknown error');
}

function toResponse(appErr) {
  const statusMap = {
    INVALID_INPUT: 400,
    UNAUTHORIZED: 401,
    RATE_LIMITED: 429,
    TRANSIENT: 503,
    INTERNAL: 500,
  };

  return {
    statusCode: statusMap[appErr.kind] ?? 500,
    body: JSON.stringify({
      ok: false,
      error: {
        kind: appErr.kind,
        code: appErr.code,
        message: appErr.message,
      },
    }),
  };
}

module.exports = {
  AppError,
  mapError,
  toResponse,
};