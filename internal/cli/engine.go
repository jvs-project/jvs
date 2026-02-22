package cli

import (
	"github.com/jvs-project/jvs/internal/engine"
	"github.com/jvs-project/jvs/pkg/model"
)

// detectEngine returns the best available engine for the repository.
func detectEngine(repoRoot string) model.EngineType {
	eng, err := engine.DetectEngine(repoRoot)
	if err != nil {
		return model.EngineCopy // fallback
	}
	return eng.Name()
}
