package pipeline

import "github.com/ztmbuilds/tavrel/internal/jsonvalue"

func resourceIdentity(attributes jsonvalue.Attributes) (service, environment string) {
	service, _ = attributes["service.name"].AsString()
	environment, _ = attributes["deployment.environment.name"].AsString()

	// Older instrumentations used deployment.environment.
	if environment == "" {
		environment, _ = attributes["deployment.environment"].AsString()
	}

	return service, environment
}
