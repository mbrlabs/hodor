package main

import (
	"fmt"
	"github.com/mbrlabs/hodor"
)

type TestMiddleware struct{}

func (m *TestMiddleware) Execute(ctx *hodor.Context) bool {
	fmt.Printf("Executing %s\n", m.Name())
	return true
}

func (m *TestMiddleware) Name() string {
	return "Test middleware"
}

func main() {
	app := hodor.NewHodor()
	app.MountBefore("", new(TestMiddleware))
	app.MountAfter("", new(TestMiddleware))

	app.Get("/test/:param", func(ctx *hodor.Context) {
		fmt.Fprintf(ctx.Writer, "get => /test/:param -> %s", ctx.UrlParams["param"])
	})

	app.Delete("/test/:param", func(ctx *hodor.Context) {
		fmt.Fprintf(ctx.Writer, "delete => /test/:param -> %s", ctx.UrlParams["param"])
	})

	app.Get("/test/hannah", func(ctx *hodor.Context) {
		fmt.Fprintf(ctx.Writer, "get => /test/hannah")
	})

	app.Post("/test/hannah", func(ctx *hodor.Context) {
		fmt.Fprintf(ctx.Writer, "post => /test/hannah")
	})

	app.Get("/test", func(ctx *hodor.Context) {
		fmt.Fprintf(ctx.Writer, "get => /test")
	})
	app.Get("/", func(ctx *hodor.Context) {
		fmt.Fprintf(ctx.Writer, "get => /")
	})

	app.Get("/static/*path", func(ctx *hodor.Context) {
		fmt.Fprintf(ctx.Writer, "get => /static/*path")
	})

	app.Post("/", func(ctx *hodor.Context) {
		fmt.Fprintf(ctx.Writer, "post => /")
	})

	app.Start()
}
