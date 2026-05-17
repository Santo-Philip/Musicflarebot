package ubot

import (
	"musicflarebot/src/vc/ntgcalls"
	"slices"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func (ctx *Context) joinPresentation(chatId int64, join bool) error {
	defer func() {
		ctx.waitConnectMu.Lock()
		if ctx.waitConnect[chatId] != nil {
			delete(ctx.waitConnect, chatId)
		}
		ctx.waitConnectMu.Unlock()
	}()
	connectionMode, err := ctx.binding.GetConnectionMode(chatId)
	if err != nil {
		return err
	}
	if connectionMode == ntgcalls.StreamConnection {
		ctx.pendingConnectionsMu.Lock()
		if ctx.pendingConnections[chatId] != nil {
			ctx.pendingConnections[chatId].Presentation = join
		}
		ctx.pendingConnectionsMu.Unlock()
	} else if connectionMode == ntgcalls.RtcConnection {
		if join {
			if !slices.Contains(ctx.presentations, chatId) {
				ctx.waitConnectMu.Lock()
				ctx.waitConnect[chatId] = make(chan error)
				ctx.waitConnectMu.Unlock()
				jsonParams, err := ctx.binding.InitPresentation(chatId)
				if err != nil {
					return err
				}
				resultParams := "{\"transport\": null}"
				inputGroupCall, err := ctx.getInputGroupCall(chatId)
				if err != nil {
					return err
				}
				callResRaw, err := ctx.App.PhoneJoinGroupCallPresentation(
					inputGroupCall,
					&tg.DataJson{
						Data: jsonParams,
					},
				)
				if err != nil {
					return err
				}
				callRes := callResRaw.(*tg.UpdatesObj)
				for _, update := range callRes.Updates {
					switch update.(type) {
					case *tg.UpdateGroupCallConnection:
						resultParams = update.(*tg.UpdateGroupCallConnection).Params.Data
					}
				}
				err = ctx.binding.Connect(
					chatId,
					resultParams,
					true,
				)
				if err != nil {
					return err
				}
				ctx.waitConnectMu.RLock()
				waitConnect := ctx.waitConnect[chatId]
				ctx.waitConnectMu.RUnlock()
				<-waitConnect
				ctx.presentations = append(ctx.presentations, chatId)
			}
		} else if slices.Contains(ctx.presentations, chatId) {
			ctx.presentations = stdRemove(ctx.presentations, chatId)
			err = ctx.binding.StopPresentation(chatId)
			if err != nil {
				return err
			}
			inputGroupCall, err := ctx.getInputGroupCall(chatId)
			if err != nil {
				return err
			}
			_, err = ctx.App.PhoneLeaveGroupCallPresentation(
				inputGroupCall,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
