import m from "mithril";
import {NButton, NInputGroup, NInput} from "./components/input";
import {NEmpty} from "./components/dataview";
import ViewJSON from "./components/ViewJSON";
import EditorJSON from "./components/EditorJSON";
import {database} from "../wailsjs/go/models";
import {use_request} from "./store";

type Request = {kind: database.Kind.REDIS} & database.RedisRequest;

export default function(id: string) {
  return {
    view() {
      // const {id} = vnode.attrs;
      // {request, response, is_loading, update_request, send}
      const r = use_request<Request, database.RedisResponse>(id);

      if (r.request === null)
        return m(NEmpty, {
          description: "Loading request...",
          class: "h100",
          style: {"justify-content": "center"},
        });

      return m("div", {
        class: "h100",
        style: {display: "grid", "grid-template-columns": "1fr 1fr", "grid-template-rows": "34px 1fr", "grid-column-gap": ".5em"},
      }, [
        m(NInputGroup, {style: {"grid-column": "span 2"}}, [
          m(NInput, {
            placeholder: "DSN",
            value: r.request.dsn,
            on: {update: (dsn: string) => r.update_request({dsn})},
          }),
          m(NButton, {
            type: "primary",
            on: {click: r.send},
            disabled: r.is_loading,
          }, "Send"),
        ]),
        m(EditorJSON, {
          class: "h100",
          value: r.request.query ?? null,
          on: {update: (value: string) => r.update_request({query: value})},
        }),
        r.response === null ?
        m(NEmpty, {
          description: "Send request or choose one from history.",
          class: "h100",
          style: {"justify-content": "center"},
        }) :
        m(ViewJSON, {value: r.response.response}),
      ]);
    },
  };
}

// <style lang="css" scoped>
// .n-tab-pane {
//   height: 100% !important;
// }
// </style>
