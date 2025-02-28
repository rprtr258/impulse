import {api} from "./api";
import type { Result} from "./result";
import {ok} from "./result";

function formatResponse(response: string): string {
  const value = ((): unknown => {
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

export async function transform(body: string, query: string): Promise<Result<string>> {
  if (query === "") {
    return ok(formatResponse(body));
  }

  const res = await api.jq(body, query);
  return res.map((item: readonly string[]) => item.map(v => formatResponse(v)).join("\n"));
};
