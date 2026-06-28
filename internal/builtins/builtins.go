package builtins

import (
	"slices"

	"github.com/albert-einshutoin/mockport/adapters/githuboauth"
	"github.com/albert-einshutoin/mockport/adapters/line"
	"github.com/albert-einshutoin/mockport/adapters/openai"
	"github.com/albert-einshutoin/mockport/adapters/slack"
	"github.com/albert-einshutoin/mockport/adapters/stripe"
	"github.com/albert-einshutoin/mockport/adapters/zohooauth"
	"github.com/albert-einshutoin/mockport/internal/adapter"
)

// builtinAdapterList is created once; adapters are stateless values.
var builtinAdapterList = []adapter.Adapter{
	stripe.New(),
	openai.New(),
	githuboauth.New(),
	slack.New(),
	line.New(),
	zohooauth.New(),
}

// Adapters returns every built-in adapter registered with the CLI.
func Adapters() []adapter.Adapter {
	return slices.Clone(builtinAdapterList)
}

// ManifestAdapters returns every built-in adapter whose compatibility claims are
// published in the support matrix. Keeping this set aligned with Adapters()
// prevents a public adapter from bypassing the checked-in manifest gate.
func ManifestAdapters() []adapter.Adapter {
	return slices.Clone(builtinAdapterList)
}
