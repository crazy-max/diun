<p align="center">
  <a href="https://pkg.go.dev/github.com/crazy-max/gohealthchecks"><img src="https://img.shields.io/badge/go.dev-docs-007d9c?logo=go&logoColor=white&style=flat-square" alt="PkgGoDev"></a>
  <a href="https://github.com/crazy-max/gohealthchecks/actions?workflow=test"><img src="https://img.shields.io/github/actions/workflow/status/crazy-max/gohealthchecks/test.yml?label=test&logo=github&style=flat-square" alt="Test workflow"></a>
  <a href="https://goreportcard.com/report/github.com/crazy-max/gohealthchecks"><img src="https://goreportcard.com/badge/github.com/crazy-max/gohealthchecks?style=flat-square" alt="Go Report"></a>
  <a href="https://codecov.io/gh/crazy-max/gohealthchecks"><img src="https://img.shields.io/codecov/c/github/crazy-max/gohealthchecks?logo=codecov&style=flat-square" alt="Codecov"></a>
  <br /><a href="https://github.com/sponsors/crazy-max"><img src="https://img.shields.io/badge/sponsor-crazy--max-181717.svg?logo=github&style=flat-square" alt="Become a sponsor"></a>
  <a href="https://www.paypal.me/crazyws"><img src="https://img.shields.io/badge/donate-paypal-00457c.svg?logo=paypal&style=flat-square" alt="Donate Paypal"></a>
</p>

## About

Go client library for accessing the [Healthchecks API](https://healthchecks.io/docs/http_api/).

## Installation

```
go get github.com/crazy-max/gohealthchecks
```

## Usage

```go
package main

import (
	"context"
	"log"
	"net/url"

	"github.com/crazy-max/gohealthchecks"
)

func main() {
	var err error
	
	// Default client uses https://hc-ping.com/
	// client := gohealthchecks.NewClient(nil)

	hcBaseURL, err := url.Parse("https://hc.foo.com")
	if err != nil {
		log.Fatal(err)
	}
	client := gohealthchecks.NewClient(&gohealthchecks.ClientOptions{
		BaseURL: hcBaseURL,
	})

	err = client.Start(context.Background(), gohealthchecks.PingingOptions{
		UUID: "5bf66975-d4c7-4bf5-bcc8-b8d8a82ea278",
		Logs: "Job started!",
	})
	if err != nil {
		log.Fatal(err)
	}

	err = client.Success(context.Background(), gohealthchecks.PingingOptions{
		UUID: "5bf66975-d4c7-4bf5-bcc8-b8d8a82ea278",
		Logs: "Job completed!",
	})
	if err != nil {
		log.Fatal(err)
	}

	err = client.Fail(context.Background(), gohealthchecks.PingingOptions{
		UUID: "5bf66975-d4c7-4bf5-bcc8-b8d8a82ea278",
		Logs: "Job failed...",
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

## Contributing

Want to contribute? Awesome! The most basic way to show your support is to star the project, or to raise issues. You
can also support this project by [**becoming a sponsor on GitHub**](https://github.com/sponsors/crazy-max) or by making
a [Paypal donation](https://www.paypal.me/crazyws) to ensure this journey continues indefinitely!

Thanks again for your support, it is much appreciated! :pray:

## License

MIT. See `LICENSE` for more details.
