package ubot

func (ctx *Context) CpuUsage() (float64, error) {
	return ctx.binding.CpuUsage()
}
