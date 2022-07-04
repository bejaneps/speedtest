# Speedtest

Speedtest is an API for testing download and upload speeds using Ookla's https://speedtest.net and Netflix https://fast.com

# Usage

First `go get` and import package

```bash
go get -v github.com/bejaneps/speedtest
```

Then, import package in your code

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bejaneps/speedtest"
)

func main() {
	measurer := speedtest.New(
		speedtest.OoklaSpeedtest,
		speedtest.WithServerCount(10),
	)

	rate, err := measurer.MeasureDownload(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rate.MbpsStr())
}
```

## TODO

* Add implementation of Upload function for Ookla's speedtest.net
* Add implementation for Netflix's fast.com tool
* Add functionality for latency check
* Add functionality for closest server
* Add warmup functionality for different length/width and workload for Ookla's speedtest.net
* Replace std logger to uber's zap
* Setup Github Action's CI for code linting and commit style check
* Add some integration tests
* Improve error messages with custom error struct
* More unit tests
