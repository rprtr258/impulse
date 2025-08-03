import m, {VnodeDOM} from "mithril";
import {EditorState} from "@codemirror/state";
import {EditorView} from "@codemirror/view";
import {PostgreSQL, sql} from "@codemirror/lang-sql";
import {defaultEditorExtensions, defaultExtensions} from "./components/editor";

type Props = {
  class?: string,
  value: string,
  on: {
    update: (value: string) => void,
  },
}

export default function() {
  let editor: EditorView | null = null;
  return {
    onupdate(vnode: VnodeDOM<Props, any>) {
      const {value, on} = vnode.attrs;

      if (editor) {
        if (value !== editor.state.doc.toString()) {
          editor.dispatch({
            changes: {
              from: 0,
              to: editor.state.doc.length,
              insert: value,
            },
          });
        }

        return;
      }

      const state = EditorState.create({
        doc: value ?? "",
        extensions: [
          ...defaultExtensions,
          ...defaultEditorExtensions(on.update),
          sql({
            dialect: PostgreSQL,
          }),
        ],
      });

      editor = new EditorView({
        parent: vnode.dom,
        state: state,
      });
    },
    view(vnode: VnodeDOM<Props, any>) {
      const props = vnode.attrs;
      return m("div", {class: props.class});
    },
  };
}
