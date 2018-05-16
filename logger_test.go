package logstsash

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/devopsfaith/krakend-gologging"
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
)

func TestNewLogger(t *testing.T) {
	logger, err := NewLogger(config.ExtraConfig{})
	if err == nil {
		t.Error("expecting an error due empty config")
		return
	}

	cfg := config.ExtraConfig{
		gologging.Namespace: map[string]interface{}{
			"level":  LEVEL_DEBUG,
			"prefix": "module_name",
			"stdout": true,
		},
	}
	logger, err = NewLogger(cfg)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	logger.Debug("mayday, mayday, mayday", map[string]interface{}{"a": true, "b": false, "cost": math.Pi})
}

func TestLogger_nothingToLog(t *testing.T) {
	buff := new(bytes.Buffer)
	l, _ := logging.NewLogger(LEVEL_DEBUG, buff, "")
	logger := Logger{
		logger:      l,
		serviceName: "some",
	}

	logger.Debug()
	logger.Debug(42)
	logger.Info()
	logger.Info(42)
	logger.Warning()
	logger.Warning(42)
	logger.Error()
	logger.Error(42)
	logger.Critical()
	logger.Critical(42)
	logger.Fatal()
	logger.Fatal(42)

	if content := buff.String(); content != "" {
		t.Errorf("unexpected log content: %s", content)
	}
}

func TestLogger(t *testing.T) {
	expectedModuleName := "module_name"
	expectedMsg := "mayday, mayday, mayday"
	buff := new(bytes.Buffer)
	l, _ := logging.NewLogger(LEVEL_DEBUG, buff, "")
	logger := Logger{
		logger:      l,
		serviceName: expectedModuleName,
	}

	location, _ := time.LoadLocation("")
	now = func() time.Time {
		return time.Unix(1526464967, 0).In(location)
	}
	defer func() { now = time.Now }()

	logger.Debug(expectedMsg, map[string]interface{}{"a": true, "b": false, "cost": 42})
	logger.Info(expectedMsg, map[string]interface{}{"a": true, "b": false, "cost": 42})
	logger.Warning(expectedMsg, map[string]interface{}{"a": true, "b": false, "cost": 42})
	logger.Error(expectedMsg, map[string]interface{}{"a": true, "b": false, "cost": 42})
	logger.Critical(expectedMsg, map[string]interface{}{"a": true, "b": false, "cost": 42})

	pattern := regexp.MustCompile("([A-Z]): ({.*})")

	fmt.Println("log content:")
	lines := strings.Split(buff.String(), "\n")
	if len(lines) < 5 {
		t.Errorf("unexpected number of logged lines (%d) : %s", len(lines), buff.String())
		return
	}
	for line, logLine := range lines[:5] {
		if !pattern.MatchString(logLine) {
			t.Errorf("The output doesn't contain the expected msg for the line %d: [%s]", line, logLine)
		}
	}
}

func TestLogger_format(t *testing.T) {
	expectedModuleName := "module_name"
	expectedMsg := "mayday, mayday, mayday"
	cfg := config.ExtraConfig{
		gologging.Namespace: map[string]interface{}{
			"level":  LEVEL_DEBUG,
			"prefix": expectedModuleName,
			"stdout": true,
		},
	}
	logger, err := NewLogger(cfg)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	location, _ := time.LoadLocation("")
	now = func() time.Time {
		return time.Unix(1526464967, 0).In(location)
	}
	defer func() { now = time.Now }()

	for i, logLevel := range []LogLevel{
		LEVEL_DEBUG,
		LEVEL_INFO,
		LEVEL_WARNING,
		LEVEL_ERROR,
		LEVEL_CRITICAL,
	} {
		data, err := logger.format(logLevel, expectedMsg, map[string]interface{}{"a": true, "b": false, "cost": 42})
		if err != nil {
			t.Errorf("unexpected error runing test case #%d: %s", i, err.Error())
			continue
		}
		var content map[string]interface{}
		if err := json.Unmarshal(data, &content); err != nil {
			t.Errorf("unexpected error unmarshaling the message #%d: %s", i, err.Error())
			continue
		}

		expectedMessage := map[string]interface{}{
			"a":          true,
			"b":          false,
			"cost":       42.0,
			"@version":   1.0,
			"@timestamp": "2018-05-16T10:02:47.000000+00:00",
			"module":     expectedModuleName,
			"host":       "localhost",
			"message":    expectedMsg,
			"level":      string(logLevel),
		}

		for k, v := range content {
			tmp, ok := expectedMessage[k]
			if !ok {
				t.Errorf("no value for key %s", k)
				continue
			}

			if !reflect.DeepEqual(tmp, v) {
				t.Errorf("unexpected value for key %s. Have: %v, Want: %v", k, v, tmp)
			}
		}
	}
}
