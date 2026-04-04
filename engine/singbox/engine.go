package singbox

import (
    "context"

    "github.com/xxf098/lite-proxy/engine"
)

// Engine is a compatibility wrapper around Runner so older call sites that
// reference singbox.Engine still work, while New() remains implemented by
// process.go.
type Engine struct {
    BinPath  string
    WorkRoot string
}

func NewEngine(binPath, workRoot string) *Engine {
    return &Engine{BinPath: binPath, WorkRoot: workRoot}
}

func (e *Engine) Name() string {
    return "sing-box"
}

func (e *Engine) Start(ctx context.Context, link string, opt engine.StartOptions) (*engine.LocalProxy, error) {
    return New(e.BinPath, e.WorkRoot).Start(ctx, link, opt)
}

var _ engine.Runner = (*Runner)(nil)
var _ engine.Runner = (*Engine)(nil)
