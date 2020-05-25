# Notifications
* [Amqp](#amqp)
* [Gotify](#gotify)
* [Mail](#mail)
* [Rocket.Chat](#rocketchat)
* [Slack](#slack)
* [Telegram](#telegram)
* [Webhook](#webhook)

## Amqp

You can send notifications to any amqp compatible server, the body will be a JSON format that looks like:

```json
{
  "diun_version": "0.3.0",
  "status": "new",
  "provider": "static-0",
  "image": "docker.io/crazymax/swarm-cronjob:0.2.1",
  "mime_type": "application/vnd.docker.distribution.manifest.v2+json",
  "digest": "sha256:5913d4b5e8dc15430c2f47f40e43ab2ca7f2b8df5eee5db4d5c42311e08dfb79",
  "created": "2019-01-24T10:26:49.152006005Z",
  "platform": "linux/amd64",
}
```

## Gotify

Notifications can be sent using a [Gotify](https://gotify.net/) instance. [Follow the instructions](https://gotify.net/docs/install) to set up a Gotify server.

![](../.res/notif-gotify.png)

## Mail

Here is an email sample if you add `mail` notification:

![](../.res/notif-mail.png)

## Rocket.Chat

To be able to send notifications to your Rocket.Chat channel, you must first create a Personal Access Token through your account settings:

![](../.res/notif-rocketchat.png)

## Slack

You can send notifications to your Slack channel using an [incoming webhook URL](https://api.slack.com/messaging/webhooks):

![](../.res/notif-slack.png)

## Telegram

Notifications can be sent via Telegram using a [Telegram Bot](https://core.telegram.org/bots).

Follow the [instructions](https://core.telegram.org/bots#6-botfather) to set up a bot and get it's token.

Message the [GetID bot](https://t.me/getidsbot) to find your chat ID.
Multiple chat IDs can be provided in order to deliver notifications to multiple recipients.

![](../.res/notif-telegram.png)

## Webhook

If you choose `webhook` notification, a HTTP request is sent with a JSON format response that looks like:

```json
{
  "diun_version": "0.3.0",
  "status": "new",
  "provider": "file",
  "image": "docker.io/crazymax/swarm-cronjob:0.2.1",
  "mime_type": "application/vnd.docker.distribution.manifest.v2+json",
  "digest": "sha256:5913d4b5e8dc15430c2f47f40e43ab2ca7f2b8df5eee5db4d5c42311e08dfb79",
  "created": "2019-01-24T10:26:49.152006005Z",
  "platform": "linux/amd64"
}
```
