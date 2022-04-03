package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"time"
)

func main() {

	g := new(errgroup.Group)
	srv1 := http.Server{
		Addr: "0.0.0.0:6060",
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		fmt.Fprintf(w, "Hello world")
	})
	srv2 := http.Server{
		Addr: "0.0.0.0:8080",
        Handler: handler,
	}
	// 启动 http server1
	g.Go(func() error {
		if err := srv1.ListenAndServe(); err != nil {
			return errors.Wrap(err, "pprof http server failed")
		}
        return nil
	})
	// 启动 http server2
	g.Go(func() error {
		if err := srv2.ListenAndServe(); err != nil {
			return errors.Wrap(err, "standard http server failed")
		}
		return nil
	})

	// 接收系统中断 信号量，退出 http 服务
	g.Go(func() error {
		s := make(chan os.Signal, 2)
		signal.Notify(s, os.Interrupt)
        <- s
		ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
		defer cancel()
		if err := srv1.Shutdown(ctx); err != nil {
			return errors.Wrap(err, "shutdown server1 err")
		}
		log.Println("has shutdown http server1")
		if err := srv2.Shutdown(ctx); err != nil {
			return errors.Wrap(err, "shutdown server2 err")
		}
		log.Println("has shutdown http server2")
		return nil
	})
	log.Println("has start http server1 listen on port 6060...")
	log.Println("has start http server2 listen on port 8080...")
    if err := g.Wait(); err != nil {
    	log.Printf("server is exiting! %v\n", err)
	}
}
