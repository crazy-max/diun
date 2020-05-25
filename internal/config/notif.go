package config

import (
	"net/mail"

	"github.com/crazy-max/diun/internal/model"
	"github.com/crazy-max/diun/pkg/utl"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
)

func (cfg *Config) validateNotif() error {
	if cfg.Notif == nil {
		return nil
	}

	if err := cfg.validateNotifAmqp(); err != nil {
		return err
	}
	if err := cfg.validateNotifGotify(); err != nil {
		return err
	}
	if err := cfg.validateNotifMail(); err != nil {
		return err
	}
	if err := cfg.validateNotifRocketChat(); err != nil {
		return err
	}
	if err := cfg.validateNotifSlack(); err != nil {
		return err
	}
	if err := cfg.validateNotifTelegram(); err != nil {
		return err
	}
	if err := cfg.validateNotifWebhook(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) validateNotifAmqp() error {
	if cfg.Notif.Amqp == nil {
		return nil
	}

	if err := mergo.Merge(cfg.Notif.Amqp, model.NotifAmqp{
		Host: "localhost",
		Port: 5672,
	}); err != nil {
		return errors.Wrap(err, "cannot set default values for amqp notif")
	}

	return nil
}

func (cfg *Config) validateNotifGotify() error {
	if cfg.Notif.Gotify == nil {
		return nil
	}

	if err := mergo.Merge(cfg.Notif.Gotify, model.NotifGotify{
		Timeout: 10,
	}); err != nil {
		return errors.Wrap(err, "cannot set default values for gotify notif")
	}

	return nil
}

func (cfg *Config) validateNotifMail() error {
	if cfg.Notif.Mail == nil {
		return nil
	}

	if _, err := mail.ParseAddress(cfg.Notif.Mail.From); err != nil {
		return errors.Wrap(err, "cannot parse sender mail address")
	}
	if _, err := mail.ParseAddress(cfg.Notif.Mail.To); err != nil {
		return errors.Wrap(err, "cannot parse recipient mail address")
	}

	if err := mergo.Merge(cfg.Notif.Mail, model.NotifMail{
		Host:               "localhost",
		Port:               25,
		SSL:                utl.NewFalse(),
		InsecureSkipVerify: utl.NewFalse(),
	}); err != nil {
		return errors.Wrap(err, "cannot set default values for mail notif")
	}

	return nil
}

func (cfg *Config) validateNotifRocketChat() error {
	if cfg.Notif.RocketChat == nil {
		return nil
	}

	if err := mergo.Merge(cfg.Notif.RocketChat, model.NotifRocketChat{
		Timeout: 10,
	}); err != nil {
		return errors.Wrap(err, "cannot set default values for rocketchat notif")
	}

	return nil
}

func (cfg *Config) validateNotifSlack() error {
	if cfg.Notif.Slack == nil {
		return nil
	}

	// noop
	return nil
}

func (cfg *Config) validateNotifTelegram() error {
	if cfg.Notif.Telegram == nil {
		return nil
	}

	// noop
	return nil
}

func (cfg *Config) validateNotifWebhook() error {
	if cfg.Notif.Webhook == nil {
		return nil
	}

	if err := mergo.Merge(cfg.Notif.Webhook, model.NotifWebhook{
		Method:  "GET",
		Timeout: 10,
	}); err != nil {
		return errors.Wrap(err, "cannot set default values for webhook notif")
	}

	return nil
}
