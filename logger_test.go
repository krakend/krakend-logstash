package logstash

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
	logger.Info()
	logger.Warning()
	logger.Error()
	logger.Critical()

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

func TestLogger_format_unexpectedMessageType(t *testing.T) {
	buff := new(bytes.Buffer)
	l, _ := logging.NewLogger(LEVEL_DEBUG, buff, "")
	logger := Logger{
		logger:      l,
		serviceName: "some",
	}

	location, _ := time.LoadLocation("")
	now = func() time.Time {
		return time.Unix(1526464967, 0).In(location)
	}
	defer func() { now = time.Now }()

	for i, testCase := range []struct {
		Expected string
		Values   []interface{}
	}{
		{
			Expected: `{"@timestamp":"2018-05-16T10:02:47.000000+00:00","@version":1,"host":"localhost","level":"DEBUG","message":"42","module":"some"}`,
			Values:   []interface{}{42},
		},
		{
			Expected: `{"@timestamp":"2018-05-16T10:02:47.000000+00:00","@version":1,"a":1,"host":"localhost","level":"DEBUG","message":"42","module":"some"}`,
			Values:   []interface{}{42, map[string]interface{}{"a": 1}},
		},
		{
			Expected: `{"@timestamp":"2018-05-16T10:02:47.000000+00:00","@version":1,"host":"localhost","level":"DEBUG","logstash.sample":{"A":1},"message":"42","module":"some"}`,
			Values:   []interface{}{42, sample{A: 1}},
		},
		{
			Expected: `{"@timestamp":"2018-05-16T10:02:47.000000+00:00","@version":1,"host":"localhost","level":"DEBUG","message":"hey there multi parts","module":"some"}`,
			Values:   []interface{}{"hey", "there", "multi", "parts"},
		},
		{
			Expected: `{"@timestamp":"2018-05-16T10:02:47.000000+00:00","@version":1,"host":"localhost","level":"DEBUG","message":"true 3 1.100000 basic types true","module":"some"}`,
			Values:   []interface{}{true, 3, 1.1, "basic types", true},
		},
		{
			Expected: `{"@timestamp":"2018-05-16T10:02:47.000000+00:00","@version":1,"host":"localhost","level":"DEBUG","logstash.sample":{"A":1},"message":"true 3 1.100000 basic types true","module":"some"}`,
			Values:   []interface{}{true, 3, sample{A: 1}, 1.1, "basic types", true},
		},
	} {
		data, err := logger.format(LEVEL_DEBUG, testCase.Values...)
		if err != nil {
			t.Errorf("unexpected error (#%d): %s", i, err.Error())
			continue
		}

		if string(data) != testCase.Expected {
			t.Errorf("unexpected result (#%d). Have: %s, Want: %s", i, string(data), testCase.Expected)
		}
	}

}

type sample struct{ A int }
