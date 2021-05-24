package logger

import (
	"fmt"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/grpclog"
)

func SetGrpcLogger(logger zerolog.Logger) {
	grpclog.SetLoggerV2(wrap(logger))
}

func wrap(l zerolog.Logger) *bridge {
	return &bridge{l}
}

type bridge struct {
	zerolog.Logger
}

func (b *bridge) Info(args ...interface{}) {
	b.Logger.Info().Msg(fmt.Sprint(args...))
}

func (b *bridge) Infoln(args ...interface{}) {
	b.Logger.Info().Msg(fmt.Sprint(args...))
}

func (b *bridge) Infof(format string, args ...interface{}) {
	b.Logger.Info().Msgf(format, args...)
}

func (b *bridge) Warning(args ...interface{}) {
	b.Logger.Warn().Msg(fmt.Sprint(args...))
}

func (b *bridge) Warningln(args ...interface{}) {
	b.Logger.Warn().Msg(fmt.Sprint(args...))
}

func (b *bridge) Warningf(format string, args ...interface{}) {
	b.Logger.Warn().Msgf(format, args...)
}

func (b *bridge) Error(args ...interface{}) {
	b.Logger.Error().Msg(fmt.Sprint(args...))
}

func (b *bridge) Errorln(args ...interface{}) {
	b.Logger.Error().Msg(fmt.Sprint(args...))
}

func (b *bridge) Errorf(format string, args ...interface{}) {
	b.Logger.Error().Msgf(format, args...)
}

func (b *bridge) Fatal(args ...interface{}) {
	b.Logger.Fatal().Msg(fmt.Sprint(args...))
}

func (b *bridge) Fatalln(args ...interface{}) {
	b.Logger.Fatal().Msg(fmt.Sprint(args...))
}

func (b *bridge) Fatalf(format string, args ...interface{}) {
	b.Logger.Fatal().Msgf(format, args...)
}

func (b *bridge) V(verbosity int) bool {
	// verbosity values:
	// 0 = info
	// 1 = warning
	// 2 = error
	// 3 = fatal
	switch b.GetLevel() {
	case zerolog.PanicLevel:
		return verbosity > 3
	case zerolog.FatalLevel:
		return verbosity == 3
	case zerolog.ErrorLevel:
		return verbosity == 2
	case zerolog.WarnLevel:
		return verbosity == 1
	case zerolog.InfoLevel:
		return verbosity == 0
	case zerolog.DebugLevel:
		return true
	case zerolog.TraceLevel:
		return true
	default:
		return false
	}
}
