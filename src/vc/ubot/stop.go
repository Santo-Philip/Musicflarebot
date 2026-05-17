package ubot

func (ctx *Context) Stop(chatId int64) error {
	ctx.presentations = stdRemove(ctx.presentations, chatId)
	ctx.callSourcesMu.Lock()
	delete(ctx.callSources, chatId)
	ctx.callSourcesMu.Unlock()
	err := ctx.binding.Stop(chatId)
	if err != nil {
		return err
	}
	ctx.inputGroupCallsMutex.RLock()
	inputGroupCall := ctx.inputGroupCalls[chatId]
	ctx.inputGroupCallsMutex.RUnlock()
	_, err = ctx.App.PhoneLeaveGroupCall(inputGroupCall, 0)
	if err != nil {
		return err
	}
	return nil
}
