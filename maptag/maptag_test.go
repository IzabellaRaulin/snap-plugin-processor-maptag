package maptag

import (
	"runtime"
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewPlugin(t *testing.T) {
	Convey("Creating new plugin", t, func() {
		plg := NewPlugin()
		So(plg, ShouldNotBeNil)
		So(plg.mapping, ShouldNotBeNil)
	})
}

func TestPlugin_GetConfigPolicy(t *testing.T) {
	plg := NewPlugin()
	Convey("Getting config policy", t, func() {
		So(func() { plg.GetConfigPolicy() }, ShouldNotPanic)

		config, err := plg.GetConfigPolicy()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)
	})
}

func TestPlugin_Process(t *testing.T) {
	Convey("Processing metrics with reftype=tag", t, func() {
		plg := NewPlugin()
		mts, err := plg.Process(mockMetrics(), mockPluginConfig_tag())
		So(err, ShouldBeNil)
		So(mts, ShouldHaveLength, 2)
		So(mts[0].Tags, ShouldContainKey, "newtag")
		So(mts[0].Tags["newtag"], ShouldEqual, "somevalue")
		So(mts[1].Tags, ShouldNotContainKey, "newtag")
	})

	Convey("Processing metrics with reftype=ns_value", t, func() {
		plg := NewPlugin()
		mts, err := plg.Process(mockMetrics(), mockPluginConfig_value())
		So(err, ShouldBeNil)
		So(mts, ShouldHaveLength, 2)
		So(mts[0].Tags, ShouldContainKey, "newtag")
		So(mts[0].Tags["newtag"], ShouldEqual, "somevalue")
		So(mts[1].Tags, ShouldContainKey, "newtag")
		So(mts[1].Tags["newtag"], ShouldEqual, "somevalue")
	})

	Convey("Processing metrics with reftype=ns_name", t, func() {
		plg := NewPlugin()
		mts, err := plg.Process(mockMetrics(), mockPluginConfig_name())
		So(err, ShouldBeNil)
		So(mts, ShouldHaveLength, 2)
		So(mts[0].Tags, ShouldNotContainKey, "newtag")
		So(mts[1].Tags, ShouldContainKey, "newtag")
		So(mts[1].Tags["newtag"], ShouldEqual, "somevalue")
	})
}

func TestPlugin_CreateCacheTTL(t *testing.T) {
	Convey("Processing metrics with data from recreated cache", t, func() {
		plg := NewPlugin()
		mts, err := plg.Process(mockMetrics(), mockPluginConfig_longTTL())
		// reset cache time
		plg.cachetime = time.Unix(0, 0)
		// put some other data into cache
		plg.mapping = map[string][]string{"first": {"valueone"}, "newtag": {"newsomevalue"}}
		// plugin should recreate cache data from because cachetime was reset
		mts, err = plg.Process(mockMetrics(), mockPluginConfig_longTTL())
		So(err, ShouldBeNil)
		So(mts, ShouldHaveLength, 2)
		So(mts[0].Tags, ShouldContainKey, "newtag")
		So(mts[0].Tags["newtag"], ShouldEqual, "somevalue")
	})
}

func TestPlugin_UseCacheTTL(t *testing.T) {
	Convey("Processing metrics with data from cache", t, func() {
		plg := NewPlugin()
		mts, err := plg.Process(mockMetrics(), mockPluginConfig_longTTL())
		// do not reset cachetime but put some other data into cache
		plg.mapping = map[string][]string{"first": {"valueone"}, "newtag": {"newsomevalue"}}
		// plugin should use cache data because of cache ttl
		mts, err = plg.Process(mockMetrics(), mockPluginConfig_longTTL())
		So(err, ShouldBeNil)
		So(mts, ShouldHaveLength, 2)
		So(mts[0].Tags, ShouldContainKey, "newtag")
		So(mts[0].Tags["newtag"], ShouldEqual, "newsomevalue")
	})
}

func mockMetrics() []plugin.Metric {
	metrics := []plugin.Metric{}
	// first
	mt := plugin.Metric{}
	mt.Data = 3.1415926
	mt.Timestamp = time.Now()
	mt.Namespace = plugin.NewNamespace("test", "static", "namespace", "pi")
	mt.Tags = map[string]string{"tagone": "valueone", "tagtwo": "valuetwo"}
	metrics = append(metrics, mt)

	// second
	mt = plugin.Metric{}
	mt.Data = 2.7182818
	mt.Timestamp = time.Now()
	mt.Namespace = plugin.NewNamespace("test")
	mt.Namespace = mt.Namespace.AddDynamicElement("dynamic", "dynamic namespace element")
	mt.Namespace = mt.Namespace.AddStaticElements("namespace", "e")
	mt.Namespace[1].Value = "valuedynamic"
	mt.Tags = map[string]string{"tagone": "anothervalueone", "tagtwo": "valuetwo"}
	metrics = append(metrics, mt)

	return metrics
}

func mockPluginConfig_tag() plugin.Config {
	cfg := plugin.Config{
		"cmd":      "?",
		"arg0":     "?",
		"arg1":     "echo valueone somevalue",
		"regex":    "(?P<first>\\S+)\\s+(?P<newtag>\\S+)",
		"reftype":  "tag",
		"refname":  "tagone",
		"refgroup": "first",
		"ttl":      int64(1),
	}
	addCmdToConfig(&cfg)
	return cfg
}

func mockPluginConfig_longTTL() plugin.Config {
	cfg := plugin.Config{
		"cmd":      "?",
		"arg0":     "?",
		"arg1":     "echo valueone somevalue",
		"regex":    "(?P<first>\\S+)\\s+(?P<newtag>\\S+)",
		"reftype":  "tag",
		"refname":  "tagone",
		"refgroup": "first",
		"ttl":      int64(1000),
	}
	addCmdToConfig(&cfg)
	return cfg
}

func mockPluginConfig_value() plugin.Config {
	cfg := &plugin.Config{
		"cmd":      "?",
		"arg0":     "?",
		"arg1":     "echo namespace somevalue",
		"regex":    "(?P<first>\\S+)\\s+(?P<newtag>\\S+)",
		"reftype":  "ns_value",
		"refname":  "namespace",
		"refgroup": "first",
		"ttl":      int64(1),
	}
	addCmdToConfig(cfg)
	return *cfg
}

func mockPluginConfig_name() plugin.Config {
	cfg := plugin.Config{
		"cmd":      "?",
		"arg0":     "?",
		"arg1":     "echo valuedynamic somevalue",
		"regex":    "(?P<first>\\S+)\\s+(?P<newtag>\\S+)",
		"reftype":  "ns_name",
		"refname":  "dynamic",
		"refgroup": "first",
		"ttl":      int64(1),
	}
	addCmdToConfig(&cfg)
	return cfg
}

func addCmdToConfig(cfg *plugin.Config) {
	switch runtime.GOOS {
	case "windows":
		(*cfg)["cmd"] = "cmd.exe"
		(*cfg)["arg0"] = "/C"
	case "linux":
		(*cfg)["cmd"] = "/bin/sh"
		(*cfg)["arg0"] = "-c"
	}
}
