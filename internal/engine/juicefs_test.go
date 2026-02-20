package engine_test

import (
	"testing"

	"github.com/jvs-project/jvs/internal/engine"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestJuiceFSEngine_Name(t *testing.T) {
	eng := engine.NewJuiceFSEngine()
	assert.Equal(t, model.EngineJuiceFSClone, eng.Name())
}

func TestJuiceFSEngine_CloneWithoutJuiceFS(t *testing.T) {
	// This test verifies the engine exists, actual clone requires juicefs command
	eng := engine.NewJuiceFSEngine()
	assert.NotNil(t, eng)
}
