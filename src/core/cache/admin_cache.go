package cache

import (
	"fmt"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
)

// AdminCache is the package-level cache for chat administrator lists.
var AdminCache = NewCache[[]*tg.Participant](time.Hour)

// adminCacheKey returns the canonical cache key for a chat's admin list.
func adminCacheKey(chatID int64) string {
	return fmt.Sprintf("admins:%d", chatID)
}

// GetChatAdminIDs returns the user IDs of all cached admins for chatID.
func GetChatAdminIDs(chatID int64) ([]int64, error) {
	admins, ok := AdminCache.Get(adminCacheKey(chatID))
	if !ok {
		return nil, fmt.Errorf("admins for chat %d not in cache", chatID)
	}

	ids := make([]int64, 0, len(admins))
	for _, admin := range admins {
		if admin.User != nil {
			ids = append(ids, admin.User.ID)
		}
	}
	return ids, nil
}

// GetAdmins returns the administrator list for chatID.
func GetAdmins(client *tg.Client, chatID int64, forceReload bool) ([]*tg.Participant, error) {
	key := adminCacheKey(chatID)

	if !forceReload {
		if admins, ok := AdminCache.Get(key); ok {
			return admins, nil
		}
	}

	admins, _, err := client.GetChatMembers(chatID, &tg.ParticipantOptions{
		Filter: &tg.ChannelParticipantsAdmins{},
	})
	if err != nil {
		return nil, fmt.Errorf("fetch admins for chat %d: %w", chatID, err)
	}

	AdminCache.Set(key, admins)
	return admins, nil
}

// GetUserAdmin returns the Participant record for userID in chatID, or an error.
func GetUserAdmin(client *tg.Client, chatID, userID int64, forceReload bool) (*tg.Participant, error) {
	admins, err := GetAdmins(client, chatID, forceReload)
	if err != nil {
		return nil, err
	}

	for _, admin := range admins {
		if admin.User != nil && admin.User.ID == userID {
			return admin, nil
		}
	}

	return nil, fmt.Errorf("user %d is not an administrator in chat %d", userID, chatID)
}

// GetRights returns the administrator rights for userID in chatID.
func GetRights(client *tg.Client, chatID, userID int64, forceReload bool) (*tg.ChatAdminRights, error) {
	admin, err := GetUserAdmin(client, chatID, userID, forceReload)
	if err != nil {
		return nil, err
	}

	if admin.Rights != nil {
		return admin.Rights, nil
	}
	if admin.Status == "creator" {
		return allCreatorRights, nil
	}

	return nil, fmt.Errorf("user %d has no admin rights in chat %d", userID, chatID)
}

// allCreatorRights is the full permission set implicitly held by a chat creator.
var allCreatorRights = &tg.ChatAdminRights{
	ChangeInfo:     true,
	DeleteMessages: true,
	EditMessages:   true,
	InviteUsers:    true,
	Other:          true,
	PinMessages:    true,
	PostMessages:   true,
	AddAdmins:      true,
	BanUsers:       true,
	ManageCall:     true,
	Anonymous:      false,
}

// ClearAdminCache removes the cached admin list for chatID.
func ClearAdminCache(chatID int64) {
	if chatID == 0 {
		AdminCache.Clear()
		return
	}
	AdminCache.Delete(adminCacheKey(chatID))
}

// UpdateAdminCache updates the cached administrator list for chatID.
func UpdateAdminCache(chatID int64, member *tg.Participant) {
	key := adminCacheKey(chatID)
	admins, ok := AdminCache.Get(key)
	if !ok {
		return
	}

	userID := int64(0)
	if member.User != nil {
		userID = member.User.ID
	}

	if userID == 0 {
		return
	}

	isAdmin := member.Status == "administrator" || member.Status == "creator"

	updated := false
	newAdmins := make([]*tg.Participant, 0, len(admins))
	for _, admin := range admins {
		currentID := int64(0)
		if admin.User != nil {
			currentID = admin.User.ID
		}

		if currentID == userID {
			if isAdmin {
				newAdmins = append(newAdmins, member)
				updated = true
			}
			continue
		}
		newAdmins = append(newAdmins, admin)
	}

	if isAdmin && !updated {
		newAdmins = append(newAdmins, member)
	}

	AdminCache.Set(key, newAdmins)
}
