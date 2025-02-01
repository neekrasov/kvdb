package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/neekrasov/kvdb/pkg/config"
	. "github.com/smartystreets/goconvey/convey" //nolint:revive,nolintlint
)

func TestConfigLoading(t *testing.T) { //nolint:funlen,nolintlint
	Convey("Given a config file", t, func() {
		tmpfile, err := os.CreateTemp("", "testconfig")
		So(err, ShouldBeNil)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.WriteString(`
engine:
  type: "in_memory"
network:
  address: "127.0.0.1:3221"
  max_connections: 200
  max_message_size: "5KB"
  idle_timeout: 6m
logging:
  level: "debug"
  output: "/log/output_test.log"
`)
		So(err, ShouldBeNil)
		err = tmpfile.Close()
		So(err, ShouldBeNil)

		var cfg config.Config
		Convey("When loading config from file", func() {
			cfg, err = config.GetConfig(tmpfile.Name())
			So(err, ShouldBeNil)
			So(cfg.Engine.Type, ShouldEqual, "in_memory")
			So(cfg.Network.Address, ShouldEqual, "127.0.0.1:3221")
			So(cfg.Network.MaxConnections, ShouldEqual, 200)
			So(cfg.Network.MaxMessageSize, ShouldEqual, "5KB")
			So(cfg.Network.IdleTimeout, ShouldEqual, 6*time.Minute)
			So(cfg.Logging.Level, ShouldEqual, "debug")
			So(cfg.Logging.Output, ShouldEqual, "/log/output_test.log")
		})

		Convey("When config file does not exist", func() {
			cfg, err = config.GetConfig("non_existent_file")
			So(err, ShouldBeNil)
			So(cfg.Engine.Type, ShouldEqual, "in_memory")
			So(cfg.Network.Address, ShouldEqual, "127.0.0.1:3223")
			So(cfg.Network.MaxConnections, ShouldEqual, 100)
			So(cfg.Network.MaxMessageSize, ShouldEqual, "4KB")
			So(cfg.Network.IdleTimeout, ShouldEqual, 5*time.Minute)
			So(cfg.Logging.Level, ShouldEqual, "info")
			So(cfg.Logging.Output, ShouldEqual, "/log/output.log")
		})
	})
}
