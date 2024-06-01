package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/kylebeee/arc53-watcher-go/errors"
)

var DefaultUserTabList = []string{
	"activity",
	"gallery",
	"links",
}

var DefaultCommunityTabsList = []string{
	"activity",
	"about",
	"collections",
	"staking",
	"subscriptions",
	"shuffles",
}

var CommunityTabListWithoutCollections = []string{
	"activity",
	"about",
	"staking",
	"subscriptions",
	"shuffles",
}

var DefaultCommunityTab = "collections"
var DefaultUserTab = "gallery"

type CommunitySettings struct {
	ID         uint64 `structs:"id,omitempty" db:"id" json:"id,omitempty"`
	DefaultTab string `structs:"default_tab,omitempty" db:"default_tab" json:"default_tab,omitempty"`
}

func CommunitySettingsTableKeys() []string {
	return []string{"id", "default_tab"}
}

func GetCommunitySettings[H Handle](h H, id uint64) (*CommunitySettings, error) {
	const op errors.Op = "GetCommunitySettings"
	query := fmt.Sprintf("select %s from %s.community_settings where id = ?", strings.Join(CommunitySettingsTableKeys(), ","), arc53Database())

	var settings CommunitySettings
	err := h.Get(&settings, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Community Settings Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &settings, nil
}

func GetAllCommunitySettings[H Handle](h H) (*CommunitySettings, error) {
	const op errors.Op = "GetAllCommunitySettings"
	query := fmt.Sprintf("select %s from %s.community_settings", strings.Join(CommunitySettingsTableKeys(), ","), arc53Database())

	var settings CommunitySettings
	err := h.Select(&settings, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(pkg, op, errors.DatabaseResultNotFound, err, "Community Settings Not Found")
		}
		return nil, errors.E(pkg, op, err)
	}

	return &settings, nil
}
