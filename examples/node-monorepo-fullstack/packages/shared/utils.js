export function formatMessage(message) {
  return `[${new Date().toISOString()}] ${message}`;
}

export function createResponse(data, status = "success") {
  return {
    status,
    data,
    timestamp: new Date().toISOString()
  };
}
