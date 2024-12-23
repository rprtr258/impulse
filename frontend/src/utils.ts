export function formatResponse(response: string): string {
  const value = (() => {
    try {
      return JSON.parse(response);
    } catch {
      return null;
    }
  })();
  if (value === null) {
    return response;
  }
  return JSON.stringify(value, null, 2);
}