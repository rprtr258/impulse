import * as monaco from "monaco-editor";

export const theme: monaco.editor.IStandaloneThemeData = {
  base: "vs-dark",
  inherit: true,
  rules: [
    {token: "number",            foreground: "F78C6C"},
    {token: "string.key.json",   foreground: "49C1AE"},
    {token: "string.value.json", foreground: "BEE28A"},
    {token: "keyword.json",      foreground: "C792EA"},
  ],
  colors: {
    'editor.background': '#0F111A',
    'editor.foreground': '#8F93A2',
    'editor.lineHighlightBackground': '#00000050',
    'editor.selectionBackground': '#717CB450',
  },
};