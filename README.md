# grouter middlewares.

<h2>a quick tutorial to using them with grouter:</h2>

```

package main

import (
	"fmt"
	"net/http"

	"github.com/72sevenzy2/grouter/grouter"
)

func main() {
	r := grouter.NewGrouter()

        r.Use(grouter.Recoverer()) // recoverer middleware must goes first if your using it, it prevents un-graceful shutdowns when the server crashes.

	limiter := grouter.NewLimiter(100, 1) // NewLimiter() takes in a maximum number of requests, and how often the token bucket is refilled.
	r.Use(limiter.RateLimiter())

	r.Use(grouter.Logger(0)) // standard logging, in which param 0 indicates default body size logging.
        // the request body read limit for logging mw can be capped, the default size being 1KiB, but can be configured via passing in an limit when using the logger, e.g. "r.Use(grouter.Logger(1024 * 2)" (2 KiB cap).

	r.Use(grouter.BearerAuth("secretKey")) // include a secret bearer auth token (can be route specific aswell).

	r.Use(grouter.BasicAuth("user1", "password1234")) // parameters username and password need to be included.

	r.Use(grouter.Timeout(time.Second * 10)) // cancels requests that last 10 seconds or more.

	r.Get("/resp", func(w http.ResponseWriter, r *grouter.Request) {
		w.Write([]byte("responded"))
	})

	fmt.Println("server running on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}

```
