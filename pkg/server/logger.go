package server

import (
	"context"

	"github.com/go-kit/kit/log"
)

type loggerMiddleware struct {
	svc    Server
	logger log.Logger
}

func (l *loggerMiddleware) PlayAudioFile(ctx context.Context, playerIP, playerPort, fileName, deviceName string) (storageUUID string, channels uint16, rate uint32, err error) {
	l.logger.Log(
		"PlayAudioFile", "start",
		"playerIP", playerIP,
		"playerPort", playerPort,
		"fileName", fileName,
		"deviceName", deviceName,
	)
	if storageUUID, channels, rate, err = l.svc.PlayAudioFile(ctx, playerIP, playerPort, fileName, deviceName); err != nil {
		l.logger.Log(
			"PlayAudioFile", "err",
			"playerIP", playerIP,
			"playerPort", playerPort,
			"fileName", fileName,
			"deviceName", deviceName,
			"err", err,
		)
	}
	l.logger.Log("PlayAudioFile", "end")
	return
}

func (l *loggerMiddleware) Play(ctx context.Context, playerIP, storageUUID, deviceName string, channels uint16, rate uint32) (err error) {
	l.logger.Log(
		"Play", "start",
		"playerIP", playerIP,
		"storageUUID", storageUUID,
		"deviceName", deviceName,
		"channels", channels,
		"rate", rate,
	)
	if err = l.svc.Play(ctx, playerIP, storageUUID, deviceName, channels, rate); err != nil {
		l.logger.Log(
			"Play", "err",
			"playerIP", playerIP,
			"storageUUID", storageUUID,
			"deviceName", deviceName,
			"channels", channels,
			"rate", rate,
			"err", err,
		)
	}
	l.logger.Log("Play", "end")
	return
}

func (l *loggerMiddleware) Pause(ctx context.Context, playerIP, deviceName string) (err error) {
	l.logger.Log(
		"Pause", "start",
		"playerIP", playerIP,
		"deviceName", deviceName,
	)
	if err = l.svc.Pause(ctx, playerIP, deviceName); err != nil {
		l.logger.Log(
			"Pause", "err",
			"playerIP", playerIP,
			"deviceName", deviceName,
			"err", err,
		)
	}
	l.logger.Log("Pause", "end")
	return
}

func (l *loggerMiddleware) Stop(ctx context.Context, playerIP, playerPort, deviceName, storageUUID string) (err error) {
	l.logger.Log(
		"Stop", "start",
		"playerIP", playerIP,
		"playerPort", playerPort,
		"deviceName", deviceName,
		"storageUUID", storageUUID,
	)
	if err = l.svc.Stop(ctx, playerIP, playerPort, deviceName, storageUUID); err != nil {
		l.logger.Log(
			"Stop", "err",
			"playerIP", playerIP,
			"playerPort", playerPort,
			"deviceName", deviceName,
			"storageUUID", storageUUID,
			"err", err,
		)
	}
	l.logger.Log("Stop", "end")
	return
}

func (l *loggerMiddleware) StartRecordingOnPlayer(ctx context.Context, playerIP, playerPort, playerDeviceName, recoderIP, recorderDeviceName string, channels, rate int) (storageUUID string, err error) {
	l.logger.Log(
		"StartRecordingOnPlayer", "start",
		"playerIP", playerIP,
		"playerPort", playerPort,
		"playerDeviceName", playerDeviceName,
		"recoderIP", recoderIP,
		"recorderDeviceName", recorderDeviceName,
		"channels", channels,
		"rate", rate,
	)
	if storageUUID, err = l.svc.StartRecordingOnPlayer(ctx, playerIP, playerPort, playerDeviceName, recoderIP, recorderDeviceName, channels, rate); err != nil {
		l.logger.Log(
			"StartRecordingInFile", "err",
			"playerIP", playerIP,
			"playerPort", playerPort,
			"playerDeviceName", playerDeviceName,
			"recoderIP", recoderIP,
			"recorderDeviceName", recorderDeviceName,
			"channels", channels,
			"rate", rate,
			"err", err,
		)
	}
	l.logger.Log("StartRecordingOnPlayer", "end")
	return
}

// NewLoggerMiddleware logger for server
func NewLoggerMiddleware(
	svc Server,
	logger log.Logger,
) Server {
	return &loggerMiddleware{
		svc:    svc,
		logger: logger,
	}
}
