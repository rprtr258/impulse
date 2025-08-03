package database

import (
	json2 "github.com/rprtr258/fun/exp/json"
)

const KindMarkdown Kind = "md"

var pluginMarkdown = plugin[MarkdownRequest, MarkdownResponse]{
	enumElem[Kind]{KindMarkdown, "MD"},
	decoderRequestMarkdown,
	decoderResponseMarkdown,
}

var decoderRequestMarkdown = json2.Map(
	func(data string) MarkdownRequest {
		return MarkdownRequest{data}
	},
	json2.Required("data", json2.String),
)

var decoderResponseMarkdown = json2.Map(
	func(data string) MarkdownResponse {
		return MarkdownResponse{data}
	},
	json2.Required("data", json2.String),
)

type MarkdownRequest struct {
	Data string `json:"data"`
}

func (MarkdownRequest) Kind() Kind { return KindMarkdown }

type MarkdownResponse struct {
	Data string `json:"data"`
}

func (MarkdownResponse) isResponseData() Kind { return KindMarkdown }
