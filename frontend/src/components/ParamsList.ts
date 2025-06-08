import m, { Vnode } from "mithril";
import {database} from '../../wailsjs/go/models';

function NDynamicInput() {
  return {
    view(vnode: Vnode<{
      value: database.KV[],
      on: {update: (value: database.KV[]) => void},
      preset: "pair",
      placeholder: {
        key: string,
        value: string,
      },
    }, any>) {
      return m("div", vnode.attrs, "not implemented");
    }
  };
}

type Props = {
  value?: database.KV[],
  on: {
    update: (value: database.KV[]) => void,
  },
};

export default function() {
  return {
    view(vnode: Vnode<Props, any>) {
      const props = vnode.attrs;

      const emit = (_: "update", value: database.KV[]) => console.log("UPDATE", value);
      // const emit = defineEmits<{
      //   update: [value: database.KV[]],
      // }>();

      const valueNonNull = props.value ?? [];

      return m(NDynamicInput, {
        value: valueNonNull.concat([{key: "", value: ""}]),
        on: {update: (value: database.KV[]) => emit("update", value)},
        preset: "pair",
        placeholder: {
          key: "Header",
          value: "Value",
        },
      });
    },
  };
}

// <style lang="css">
// div.n-dynamic-input-preset-pair > div:nth-child(1) {
//   margin-right: 4px !important;
// }
// </style>
