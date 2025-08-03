import m, {VnodeDOM} from "mithril";
import {ChangeSpec, EditorState} from "@codemirror/state";
import {EditorView} from "@codemirror/view";
import {json} from "@codemirror/lang-json";
import {defaultEditorExtensions, defaultExtensions} from "./editor";

type Props = {
  value: string | null,
  on: {
    update: (value: string) => void,
  },
  class?: string,
  style?: any,
};

export default function() {
  let editor: EditorView | null = null;
  return {
    oncreate(vnode: VnodeDOM<Props, any>) {
      const props = vnode.attrs;

      const state = EditorState.create({
        doc: props.value ?? "",
        extensions: [
          ...defaultExtensions,
          ...defaultEditorExtensions(props.on.update),
          json(),
        ],
      });

      editor = new EditorView({
        parent: vnode.dom,
        state: state,
      });

      if (props.value !== editor.state.doc.toString()) {
        editor.dispatch({
          changes: {
            from: 0,
            to: editor.state.doc.length,
            insert: props.value,
          } as ChangeSpec,
        });
      }
      props.on.update(state.doc.toString());
    },
    onremove() {
      editor?.destroy();
    },
    view(vnode: VnodeDOM<Props, any>) {
      const props = vnode.attrs;
      return m("div", {
        class: props.class,
        style: props.style ?? {},
      });
    },
  };
}
