package configembed

import "embed"

// FS bundles the default SCIM 2.0 schemas, resource types, bucket settings, and
// service provider configuration into the compiled GoSCIM binary so standalone
// installs (`goscim -demo`, `SCIM_STORE=sqlite goscim`) work from any directory
// without external files.
//
//go:embed resourceType/*.json schemas/*.json serviceProviderConfig/*.json bucketSettings/*.json
var FS embed.FS
