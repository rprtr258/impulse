import m, {Vnode} from "mithril";
import {NInput, NButton, NInputGroup} from "./components/input";
import {NEmpty} from "./components/dataview";
import ViewJSON from "./components/ViewJSON";
import EditorJSON from "./components/EditorJSON";
import {database} from "../wailsjs/go/models";
import {use_request} from "./store";

type Request = {kind: database.Kind.JQ} & database.JQRequest;
type Props = {
  id: string,
};

export default function() {
  return {
    view(vnode: Vnode<Props, any>) {
      const {id} = vnode.attrs;

      // {request, response, is_loading, update_request, send}
      const r = use_request<Request, database.JQResponse>(id);

      const jqerror: string | null = null; // TODO: use
      const responseText = (r.response?.response ?? []).join("\n");

      if (r.request === null) {
        return m(NEmpty, {
          description: "Loading request...",
          class: "h100",
          style: { "justify-content": "center" },
        });
      }

      return m("div", {
        class: "h100",
        style: {display: "grid", "grid-template-columns": "1fr 1fr", "grid-template-rows": "34px 1fr", "grid-column-gap": ".5em"},
      }, [
        m(NInputGroup, {style: {"grid-column": "span 2"}}, [
          m(NInput, {
            placeholder: "JQ query",
            status: jqerror !== null ? "error" : "success",
            value: r.request.query,
            on: {update: (query: string) => r.update_request({query})},
          }),
          // TODO: autosend
          m(NButton, {
            type: "primary",
            on: {click: r.send},
            disabled: r.is_loading,
          }, "Send"),
        ]),
        m(EditorJSON, {
          class: "h100",
          value: r.request.json,
          on: {update: (json: string) => r.update_request({json})},
        }),
        jqerror !== null ?
        m("div", {style: {position: "fixed", color: "red", bottom: "3em"}}, jqerror) :
        r.response === null ?
        m(NEmpty, {
          description: "Send request or choose one from history.",
          class: "h100",
          style: {"justify-content": "center"},
        }) :
        m(ViewJSON, {value: responseText}),
      ]);
    },
  };
}

// <style lang="css" scoped>
// .n-tab-pane {
//   height: 100% !important;
// }
// </style>
