package input

import (
	"context"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/inbox"
)

type Inbox interface {
	// List answers a query of the inbox as seen by a staff profile.
	List(ctx context.Context, viewer definition.Profile, q inbox.Query) (inbox.List, error)
}
