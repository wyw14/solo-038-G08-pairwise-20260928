package color

import (
	"github.com/wyw14/solo-038/internal/journal"
	"github.com/wyw14/solo-038/internal/model"
)

func Releases(store *journal.Store) ([]journal.ColorReleaseRecord, error) {
	events, err := store.Events()
	if err != nil {
		return nil, err
	}
	releases := []journal.ColorReleaseRecord{}
	for _, event := range events {
		if event.Kind != "color_released" {
			continue
		}
		record, err := model.DecodeEventPayload[journal.ColorReleaseRecord](event)
		if err != nil {
			return nil, err
		}
		if record.Released {
			releases = append(releases, record)
		}
	}
	return releases, nil
}
