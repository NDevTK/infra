package client

import (
	"context"

	"infra/monitoring/messages"

	gfipb "go.chromium.org/luci/bisection/proto"
)

// CrBug returns bug information.
type CrBug interface {
	// CrBugItems returns issue matching label.
	CrbugItems(ctx context.Context, label string) ([]messages.CrbugItem, error)
}

// FindIt returns FindIt information.
type FindIt interface {
	// FinditBuildbucket returns FindIt results for a build. Both input and output are using buildbucket concepts.
	FinditBuildbucket(ctx context.Context, buildID int64, failedSteps []string) ([]*messages.FinditResultV2, error)
}

// GoFindit returns information about failures that GoFindit analyzes
type GoFindit interface {
	QueryGoFinditResults(c context.Context, bbid int64, stepName string) (*gfipb.QueryAnalysisResponse, error)
}

// CrRev returns redirects for commit positions.
type CrRev interface {
	// GetRedirect gets the redirect for a commit position.
	GetRedirect(c context.Context, pos string) (map[string]string, error)
}
