package types

import (
	"github.com/chromedp/cdproto/cdp"
)

type PageNode struct {
	Node *cdp.Node
	Page int
	Url  string
}
