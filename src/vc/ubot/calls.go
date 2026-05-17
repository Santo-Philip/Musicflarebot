package ubot

import "musicflarebot/src/vc/ntgcalls"

func (ctx *Context) Calls() map[int64]*ntgcalls.CallInfo {
	return ctx.binding.Calls()
}
