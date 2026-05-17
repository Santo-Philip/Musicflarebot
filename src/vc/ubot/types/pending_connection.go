package types

import "musicflarebot/src/vc/ntgcalls"

type PendingConnection struct {
	MediaDescription ntgcalls.MediaDescription
	Payload          string
	Presentation     bool
}
