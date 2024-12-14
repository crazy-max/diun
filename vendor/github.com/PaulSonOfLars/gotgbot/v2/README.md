# Golang Telegram Bot library

Heavily inspired by the [python-telegram-bot library](https://github.com/python-telegram-bot/python-telegram-bot), this
package is a code-generated wrapper for the telegram bot api. We also provide an extensions package which defines an
updater/dispatcher pattern to provide update processing out of the box.

All the telegram types and methods are generated from
[a bot api spec](https://github.com/PaulSonOfLars/telegram-bot-api-spec). These are generated in the `gen_*.go` files.

If you have any questions, come find us in our [telegram support chat](https://t.me/GotgbotChat)!

## Features:

- All telegram API types and methods are generated from the bot api docs, which makes this library:
    - Guaranteed to match the docs
    - Easy to update
    - Self-documenting (Re-uses pre-existing telegram docs)
- Type safe; no weird interface{} logic, all types match the bot API docs.
- No third party library bloat; only uses standard library.
- Updates are each processed in their own go routine, encouraging concurrent processing, and keeping your bot
  responsive.
- Code panics are automatically recovered from and logged, avoiding unexpected downtime.

## Getting started

Download the library with the standard `go get` command:

```bash
go get github.com/PaulSonOfLars/gotgbot/v2
```

### Example bots

Sample bots can be found in the [samples directory](./samples).

Some recommended bots to look at:

- [Command Bot](./samples/commandBot): To explore basic use of commands
- [Webhook Bot](./samples/echoWebhookBot): To set up webhooks
- [Stateful Client Bot](./samples/statefulClientBot): To pass around shared data without global variables

## Docs

Docs can be found [here](https://pkg.go.dev/github.com/PaulSonOfLars/gotgbot/v2).

## Contributing

Contributions are welcome! More information on contributing can be found [here](.github/CONTRIBUTING.md).

### Regenerating the generated code.

If you've made changes to the code generation, you will probably need to regenerate the library code.
This can be done simply by running `go generate` from the repo root. Running this will generate the code from the
specification repo at the commit pinned in the `spec_commit` file.

To upgrade the commit in `spec_commit` and regenerate your code, simply run `GOTGBOT_UPGRADE=true go generate`.
This will fetch the latest commit sha, and regenerate the library against that, giving you the latest version
available.
