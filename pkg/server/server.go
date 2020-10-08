package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"sync"
)

// errors
var (
	ErrDeviceIsBusy   = errors.New("device is busy")
	ErrDeviceNotFound = errors.New("device not found")
	ErrPortIsBusy     = errors.New("port is busy")
	ErrPortNotFound   = errors.New("port not found")
)

type audio interface {
	Reader(data []byte) (io.Reader, uint16, uint32, error)
	Writer(fileName string, channels uint16, rate uint32) (io.WriteCloser, error)
}

type udp interface {
	Send(ctx context.Context, host string, r io.Reader) error
	Receive(ctx context.Context, receivePort string, w io.Writer) (err error)
}

type player interface {
	ReceiveStart(ctx context.Context, ip, port string, uuid *string) (sUUID string, err error)
	ReceiveStop(ctx context.Context, ip, port string) (err error)
	Play(ctx context.Context, ip, uuid, deviceName string, channels, rate uint32) (err error)
	Stop(ctx context.Context, ip, deviceName string) (err error)
	ClearStorage(ctx context.Context, ip, uuid string) (err error)
}

type recorder interface {
	Start(ctx context.Context, destAddr, recorderIP, deviceName string, channels, rate uint32) (err error)
	Stop(ctx context.Context, recorderIP, deviceName string) (err error)
}

// Server ...
type Server interface {
	FilePlaying(ctx context.Context, file, playerIP, playerPort, playerDeviceName string) (uuid string, channels uint16, rate uint32, err error)

	PlayerReceiveStart(ctx context.Context, playerIP, playerPort string, uuid *string) (string, error)
	PlayerReceiveStop(ctx context.Context, playerIP, playerPort string) error
	PlayerPlay(ctx context.Context, playerIP, uuid, playerDeviceName string, channels, rate uint32) (err error)
	PlayerPause(ctx context.Context, playerIP, playerDeviceName string) (err error)
	PlayerStop(ctx context.Context, playerIP, playerPort, playerDeviceName, uuid string) (err error)

	StartFileRecoding(ctx context.Context, recorderIP, recorderDeviceName string, channels, rate uint32, receivePort, file string) (err error)
	StopFileRecoding(ctx context.Context, recorderIP, recorderDeviceName, receivePort string) error
	PlayFromRecorder(ctx context.Context, playerIP, playerPort, playerDeviceName string, channels, rate uint32, recorderIP, recorderDeviceName string) (uuid string, err error)
	StopFromRecorder(ctx context.Context, playerIP, playerPort, playerDeviceName, uuid, recorderIP, recorderDeviceName string) error

	RecorderStart(ctx context.Context, recorderIP, recorderDeviceName string, channels, rate uint32, dstAddr string) error
	RecoderStop(ctx context.Context, recorderIP, recorderDeviceName string) error
}

type server struct {
	mutexSending sync.Mutex
	sending      map[string]func()

	mutexReceiving sync.Mutex
	receiving      map[string]func()

	audio    audio
	player   player
	recorder recorder
	udp      udp

	serverIP     string
	addrLayout   string
	deviceLayout string
}

// FilePlaying send file to player with playerIP on port and play on playerDeviceName
// channel and rate audio info from file
// player save audio from server in storage with uuid
func (s *server) FilePlaying(ctx context.Context, file, playerIP, playerPort, playerDeviceName string) (uuid string, channels uint16, rate uint32, err error) {
	var data []byte
	if data, err = ioutil.ReadFile(file); err != nil {
		return
	}
	var r io.Reader
	if r, channels, rate, err = s.audio.Reader(data); err != nil {
		return
	}

	if uuid, err = s.PlayerReceiveStart(ctx, playerIP, playerPort, nil); err != nil {
		return
	}

	if err = s.startSending(ctx, playerIP, playerPort, r); err != nil {
		return
	}

	if err = s.PlayerPlay(ctx, playerIP, uuid, playerDeviceName, uint32(channels), rate); err != nil {
		s.PlayerReceiveStop(ctx, playerIP, playerPort)
		s.stopSending(ctx, playerIP, playerPort)
	}
	return
}

// PlayerReceiveStart player with playerIP start receive signal from server on playerPort.
// uuid of the storage existing on the player
// if the storage with uuid does not exist or the uuid is nil, a new storage will be created on the player
// The signal will be stored in the storage sUUID
func (s *server) PlayerReceiveStart(ctx context.Context, playerIP, playerPort string, uuid *string) (sUUID string, err error) {
	return s.player.ReceiveStart(ctx, playerIP, playerPort, uuid)
}

// PlayerReceiveStop player with playerIP stop receive signal from server on playerPort.
func (s *server) PlayerReceiveStop(ctx context.Context, playerIP, playerPort string) error {
	return s.player.ReceiveStop(ctx, playerIP, playerPort)
}

// PlayerPlay play audio from storage with uuid on player with playerIP on playerDeviceName
// channels, rate - params audio
func (s *server) PlayerPlay(ctx context.Context, playerIP, uuid, playerDeviceName string, channels, rate uint32) (err error) {
	return s.player.Play(ctx, playerIP, uuid, playerDeviceName, channels, rate)
}

// PlayerPause pause audio on player with playerIP on playerDeviceName
func (s *server) PlayerPause(ctx context.Context, playerIP, playerDeviceName string) (err error) {
	return s.player.Stop(ctx, playerIP, playerDeviceName)
}

// PlayerStop stop audio from storage uuid on player with playerIP on playerDeviceName
func (s *server) PlayerStop(ctx context.Context, playerIP, playerPort, playerDeviceName, uuid string) (err error) {
	s.stopSending(ctx, playerIP, playerPort)
	s.player.ReceiveStop(ctx, playerIP, playerPort)
	s.PlayerPause(ctx, playerIP, playerDeviceName)
	s.player.ClearStorage(ctx, playerIP, uuid)
	return
}

// StartFileRecoding start receive on receivePort audio signal from recorder with recorderIP from recordeDeviceName and write in file
// channels, rate - params audio
func (s *server) StartFileRecoding(ctx context.Context, recorderIP, recorderDeviceName string, channels, rate uint32, receivePort, file string) (err error) {
	var wc io.WriteCloser
	if wc, err = s.audio.Writer(file, uint16(channels), rate); err != nil {
		return
	}
	if err = s.startReceive(ctx, recorderIP, receivePort, wc); err != nil {
		return
	}

	receiveAddr := fmt.Sprintf(s.addrLayout, s.serverIP, receivePort)
	if err = s.RecorderStart(ctx, recorderIP, recorderDeviceName, channels, rate, receiveAddr); err != nil {
		s.stopReceive(ctx, receivePort)
	}
	return
}

// StopFileRecoding stop receive on receivePort audio signal from recorder with recorderIP from recordeDeviceName
func (s *server) StopFileRecoding(ctx context.Context, recorderIP, recorderDeviceName, receivePort string) error {
	s.stopReceive(ctx, receivePort)
	s.recorder.Stop(ctx, recorderIP, recorderDeviceName)
	return nil
}

// PlayFromRecorder play audio on player with playerIP from recorder with recorderIP
func (s *server) PlayFromRecorder(ctx context.Context, playerIP, playerPort, playerDeviceName string, channels, rate uint32, recorderIP, recorderDeviceName string) (uuid string, err error) {
	if uuid, err = s.PlayerReceiveStart(ctx, playerIP, playerPort, nil); err != nil {
		return
	}

	if err = s.PlayerPlay(ctx, playerIP, uuid, playerDeviceName, channels, rate); err != nil {
		s.PlayerReceiveStop(ctx, playerIP, playerPort)
		return
	}

	dstAddr := fmt.Sprintf(s.addrLayout, playerIP, playerPort)
	if err = s.RecorderStart(ctx, recorderIP, recorderDeviceName, channels, rate, dstAddr); err != nil {
		s.PlayerReceiveStop(ctx, playerIP, playerPort)
		s.PlayerStop(ctx, playerIP, playerPort, playerDeviceName, uuid)
	}
	return
}

// StopFromRecorder stop audio on player with playerIP from recorder with recorderIP
func (s *server) StopFromRecorder(ctx context.Context, playerIP, playerPort, playerDeviceName, uuid, recorderIP, recorderDeviceName string) error {
	s.PlayerReceiveStop(ctx, playerIP, playerPort)
	s.PlayerStop(ctx, playerIP, playerPort, playerDeviceName, uuid)
	s.RecoderStop(ctx, recorderIP, recorderDeviceName)
	return nil
}

// RecorderStart start recording audio on recorder with recorderIP from recorderDeviceName and receive on dstAddr
// channels, rate - recoding param
func (s *server) RecorderStart(ctx context.Context, recorderIP, recorderDeviceName string, channels, rate uint32, dstAddr string) error {
	return s.recorder.Start(ctx, dstAddr, recorderIP, recorderDeviceName, channels, rate)
}

// RecoderStop stop recording audio on recorder with recorderIP from recorderDeviceName
func (s *server) RecoderStop(ctx context.Context, recorderIP, recorderDeviceName string) error {
	return s.recorder.Stop(ctx, recorderIP, recorderDeviceName)
}

func (s *server) startSending(ctx context.Context, playerIP, playerPort string, r io.Reader) (err error) {
	s.mutexSending.Lock()
	defer s.mutexSending.Unlock()

	dstAddr := fmt.Sprintf(s.addrLayout, playerIP, playerPort)
	if _, isExist := s.sending[dstAddr]; !isExist {
		c, stop := context.WithCancel(context.Background())
		if err = s.udp.Send(c, dstAddr, r); err == nil {
			s.sending[dstAddr] = stop
			return
		}
		stop()
		return
	}
	err = ErrDeviceIsBusy
	return
}

func (s *server) stopSending(ctx context.Context, playerIP, playerPort string) (err error) {
	s.mutexSending.Lock()
	defer s.mutexSending.Unlock()

	dstAddr := fmt.Sprintf(s.addrLayout, playerIP, playerPort)
	if stop, isExist := s.sending[dstAddr]; isExist {
		stop()
		delete(s.sending, dstAddr)
		return
	}
	err = ErrDeviceNotFound
	return
}

func (s *server) startReceive(ctx context.Context, recorderIP, receivePort string, wc io.WriteCloser) (err error) {
	s.mutexReceiving.Lock()
	defer s.mutexReceiving.Unlock()

	if _, isExist := s.receiving[receivePort]; !isExist {
		c, stop := context.WithCancel(context.Background())
		if err = s.udp.Receive(c, receivePort, wc); err == nil {
			s.receiving[receivePort] = func() {
				stop()
				wc.Close()
			}
			return
		}
		stop()
		return
	}
	err = ErrPortIsBusy
	return
}

func (s *server) stopReceive(ctx context.Context, receivePort string) (err error) {
	s.mutexReceiving.Lock()
	defer s.mutexReceiving.Unlock()

	if stop, isExist := s.receiving[receivePort]; isExist {
		stop()
		delete(s.receiving, receivePort)
		return
	}
	err = ErrPortNotFound
	return
}

// NewServer ...
func NewServer(
	audio audio,
	recorder recorder,
	player player,
	udp udp,

	serverIP string,
	addrLayout string,
	deviceLayout string,
) Server {
	return &server{
		receiving: make(map[string]func()),
		sending:   make(map[string]func()),

		audio:    audio,
		recorder: recorder,
		player:   player,
		udp:      udp,

		serverIP:     serverIP,
		addrLayout:   addrLayout,
		deviceLayout: deviceLayout,
	}
}
