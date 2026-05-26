package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/config"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/stretchr/testify/require"
)

func TestNewHealthchecksClientResolvesUUIDFile(t *testing.T) {
	const uuid = "5bf66975-d4c7-4bf5-bcc8-b8d8a82ea278"

	uuidFile := filepath.Join(t.TempDir(), "uuid")
	require.NoError(t, os.WriteFile(uuidFile, []byte(uuid), 0o600))

	hc, err := newHealthchecksClient(&model.Healthchecks{
		UUIDFile: uuidFile,
	})

	require.NoError(t, err)
	require.NotNil(t, hc.Client)
	require.Equal(t, uuid, hc.UUID)
}

func TestStartReturnsImmediatelyWithoutSchedule(t *testing.T) {
	diun := newTestDiun(t, "")

	errCh := make(chan error, 1)
	go func() {
		errCh <- diun.Start(context.Background())
	}()

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Start to return without a schedule")
	}
}

func TestStartReturnsWhenContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	t.Cleanup(func() { cancel(nil) })

	diun := newTestDiun(t, "@every 1m")

	errCh := make(chan error, 1)
	go func() {
		errCh <- diun.Start(ctx)
	}()

	require.Never(t, func() bool {
		select {
		case err := <-errCh:
			require.NoError(t, err)
			return true
		default:
			return false
		}
	}, 100*time.Millisecond, 10*time.Millisecond)

	cancel(nil)

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Start to return")
	}
}

func TestStartReturnsWhenContextAlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(nil)

	require.NoError(t, (&Diun{}).Start(ctx))
}

func newTestDiun(t *testing.T, schedule string) *Diun {
	t.Helper()

	watch := (&model.Watch{}).GetDefaults()
	watch.RunOnStartup = new(false)
	watch.Schedule = schedule

	return newTestDiunWithWatch(t, watch)
}

func newTestDiunWithWatch(t *testing.T, watch *model.Watch) *Diun {
	t.Helper()

	diun, err := New(model.Meta{
		ID:      "diun",
		Name:    "Diun",
		Version: "test",
	}, &config.Config{
		Db: &model.Db{
			Path: filepath.Join(t.TempDir(), "diun.db"),
		},
		Watch:     watch,
		Defaults:  (&model.Defaults{}).GetDefaults(),
		Providers: &model.Providers{File: &model.PrdFile{}},
	}, "127.0.0.1:0")
	require.NoError(t, err)
	return diun
}
