package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/geoirb/sound-ethernet-streaming/pkg/recoder"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/kelseyhightower/envconfig"

	"github.com/geoirb/sound-ethernet-streaming/pkg/player"
	"github.com/geoirb/sound-ethernet-streaming/pkg/server"
	"github.com/geoirb/sound-ethernet-streaming/pkg/udp"
	"github.com/geoirb/sound-ethernet-streaming/pkg/wav"
)

type configuration struct {
	PlayerPort  string `envconfig:"PLAYER_PORT" default:"8081"`
	RecoderPort string `envconfig:"RECODER_PORT" default:"8082"`
	UDPBuffSize int    `envconfig:"UDP_BUF_SIZE" default:"1024"`
	HostLayout  string `envconfig:"HOST_LAYOUT" default:"%s:%s"`
	File        string `envconfig:"FILE" default:"/home/geo/go/src/github.com/geoirb/sound-ethernet-streaming/audio/test.wav"`
}

func main() {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	level.Info(logger).Log("msg", "initializing")

	var (
		err error
		cfg configuration
	)
	if err = envconfig.Process("", &cfg); err != nil {
		level.Error(logger).Log("msg", "failed to load configuration", "err", err)
		os.Exit(1)
	}

	wav := wav.NewWAV()
	player := player.NewClient(
		cfg.HostLayout,
		cfg.PlayerPort,
	)
	recoder := recoder.NewClient(
		cfg.HostLayout,
		cfg.RecoderPort,
	)
	udp := udp.NewUDP(cfg.UDPBuffSize)
	server := server.NewServer(
		wav,
		recoder,
		player,
		udp,

		cfg.HostLayout,
	)
	fmt.Println(server.AddFilePlayer(context.Background(), "127.0.0.1", "8083", "hw:1,0", cfg.File))

	level.Error(logger).Log("msg", "server start")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	level.Error(logger).Log("msg", "received signal, exiting signal", "signal", <-c)

	fmt.Println(server.DeletePlayer(context.Background(), "127.0.0.1", "8083"))
}
