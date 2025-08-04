import m from "mithril";
import {NEmpty} from "./components/dataview";
import EditorJSON from "./components/EditorJSON";
import {database} from "../wailsjs/go/models";
import {use_request} from "./store";
import {api} from "./api";

type Request = {kind: database.Kind.MD} & database.MarkdownRequest;

export default function(
  id: string,
  show_request: () => boolean,
): m.ComponentTypes<any, any> {
  return {
    view() {
      // {request, response, is_loading, update_request, send}
      const r = use_request<Request, database.MarkdownResponse>(id);

      let Markdownerror: string | null = null; // TODO: use
      const responseText = r.response?.data ?? "";

      function update_request(data: string): void {
        r.update_request({data});
        api.requestPerform(id)
          .then(res => {
            if (res.kind === "err") {
              Markdownerror = `Could not render document: ${res.value}`;
            } else {
              Markdownerror = null;
              r.response = res.value.response as database.MarkdownResponse;
            }
          }).
          then(() => m.redraw());
      }
      if (r.request === null) {
        return m(NEmpty, {
          description: "Loading request...",
          class: "h100",
          style: {"justify-content": "center"},
        });
      }

      return m("div", {
        class: "h100",
        style: {
          display: "grid",
          "grid-template-columns": "1fr" + (show_request() ? " 1fr" : ""),
          "grid-column-gap": ".5em",
        },
      }, [
        show_request() && m(EditorJSON, { // TODO: editor for markdown
          value: r.request.data,
          on: {update: (data: string) => update_request(data)},
          class: "h100",
          style: {
            "overflow-y": "scroll",
          },
        }),
        Markdownerror !== null ?
        m("div", {style: {position: "fixed", color: "red", bottom: "3em"}}, Markdownerror) :
        r.response === null ?
        m(NEmpty, {
          description: "Send request or choose one from history.",
          class: "h100",
          style: {"justify-content": "center"},
        }) :
        m("div", {
          class: "h100 markdown-body",
          style: {
            "overflow-y": "scroll",
          },
        }, m.trust(responseText)),
      ]);
    },
  };
}