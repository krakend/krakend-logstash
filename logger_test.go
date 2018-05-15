package logstsash

import (
	"math"
	"testing"

	"github.com/devopsfaith/krakend-gologging"
	"github.com/devopsfaith/krakend/config"
)

func TestLogger_format(t *testing.T) {
	cfg := config.ExtraConfig{
		gologging.Namespace: map[string]interface{}{
			"level":  LEVEL_DEBUG,
			"prefix": "module_name",
			"stdout": true,
		},
	}
	logger, err := NewLogger(cfg)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	logger.Debug("mayday, mayday, mayday", map[string]interface{}{"a": true, "b": false, "cost": math.Pi})
	logger.Info("mayday, mayday, mayday", map[string]interface{}{"a": true, "b": false, "cost": math.Pi})
	logger.Warning("mayday, mayday, mayday", map[string]interface{}{"a": true, "b": false, "cost": math.Pi})
	logger.Error("mayday, mayday, mayday", map[string]interface{}{"a": true, "b": false, "cost": math.Pi})
	logger.Critical("mayday, mayday, mayday", map[string]interface{}{"a": true, "b": false, "cost": math.Pi})
}
