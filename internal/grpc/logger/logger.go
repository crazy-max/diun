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

func (b *bridge) Info(args ...any) {
	b.Logger.Info().Msg(fmt.Sprint(args...))
}

func (b *bridge) Infoln(args ...any) {
	b.Logger.Info().Msg(fmt.Sprint(args...))
}

func (b *bridge) Infof(format string, args ...any) {
	b.Logger.Info().Msgf(format, args...)
}

func (b *bridge) Warning(args ...any) {
	b.Logger.Warn().Msg(fmt.Sprint(args...))
}

func (b *bridge) Warningln(args ...any) {
	b.Logger.Warn().Msg(fmt.Sprint(args...))
}

func (b *bridge) Warningf(format string, args ...any) {
	b.Logger.Warn().Msgf(format, args...)
}

func (b *bridge) Error(args ...any) {
	b.Logger.Error().Msg(fmt.Sprint(args...))
}

func (b *bridge) Errorln(args ...any) {
	b.Logger.Error().Msg(fmt.Sprint(args...))
}

func (b *bridge) Errorf(format string, args ...any) {
	b.Logger.Error().Msgf(format, args...)
}

func (b *bridge) Fatal(args ...any) {
	b.Logger.Fatal().Msg(fmt.Sprint(args...))
}

func (b *bridge) Fatalln(args ...any) {
	b.Logger.Fatal().Msg(fmt.Sprint(args...))
}

func (b *bridge) Fatalf(format string, args ...any) {
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
