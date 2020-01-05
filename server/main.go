package main

import (
	"context"
	"fmt"
	"log"
	"main/config"
	"main/manager"
	"main/pubsub"
	"main/pubsub/systemevent"
	"main/router"
	"main/transport"
	"net/http"
	"net/http/httputil"
	"os"
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	"golang.org/x/crypto/acme/autocert"

	_ "main/statik"

	_ "github.com/mattn/go-sqlite3"

	"github.com/labstack/echo/v4"
)

// NewMultipleHostReverseProxy creates a reverse proxy that will randomly
// select a host from the passed `targets`
func NewMultipleHostReverseProxy(conf config.Config) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		for _, target := range conf.Targets {
			if req.Host == target.Host {
				req.URL.Scheme = "http"
				req.URL.Host = target.Proxy
				log.Printf("proxy to %s\n", target.Proxy)
				return
			}
		}

		for _, target := range conf.Targets {
			if target.Default {
				req.URL.Scheme = "http"
				req.URL.Host = target.Proxy
				log.Printf("proxy to %s\n", target.Proxy)
				return
			}
		}

		req.URL.Scheme = "http"
		req.URL.Host = "localhost:8080" // webuiの portに合わせる
		req.URL.Path = "/api/404"
	}

	return &httputil.ReverseProxy{Director: director}
}

func getDatabase() (*gorm.DB, error) {
	engine, err := gorm.Open("sqlite3", "proxy.db")

	if err != nil {
		return nil, err
	}

	return engine, nil
}

func main() {
	fmt.Println("poi")
	os.Mkdir("repository", 0777)
	engine, err := getDatabase()
	if err != nil {
		panic(err)
	}

	pubsub.SystemEvent.Pub(pubsub.System{Time: time.Now(), Type: systemevent.SERVER_START})

	manager := manager.New(engine)

	// 設定再読み込みなどなどを行う
	var resetF func() = func() {}
	pubsub.UpdateConfigEvent.Sub(func(event pubsub.UpdateConfig) {
		// ちゃんとロックを取らないとヤバそう
		resetF()
		conf := manager.Config.Get()

		httpsHosts := make([]string, 0)
		proxy := NewMultipleHostReverseProxy(conf)
		proxy.Transport = transport.New()
		pubsub.SystemEvent.Pub(pubsub.System{Time: time.Now(), Type: systemevent.DIRECTORS_REGISTER})

		httpsSrv := &http.Server{Handler: proxy}
		httpSrv := &http.Server{Addr: ":80", Handler: proxy}

		for _, target := range conf.Targets {
			if target.Https {
				httpsHosts = append(httpsHosts, target.Host)
			}
			if target.Repository != "" {
				pubsub.GetWebhookEvent.Pub(pubsub.GetWebhook{Repository: target.Repository})
			}
		}
		pubsub.SystemEvent.Pub(pubsub.System{Time: time.Now(), Type: systemevent.NEW_SETTINGS_APPLY})

		go func() {
			httpsSrv.Serve(autocert.NewListener(httpsHosts...))
		}()
		go func() {
			httpSrv.ListenAndServe()
		}()

		resetF = func() {
			wg := sync.WaitGroup{}
			wg.Add(2)
			go func() {
				ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
				if err := httpsSrv.Shutdown(ctx); err != nil {
					log.Fatal(err)
				}
				wg.Done()
			}()

			go func() {
				ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
				if err := httpSrv.Shutdown(ctx); err != nil {
					log.Fatal(err)
				}
				wg.Done()
			}()
			wg.Wait()

		}
	})

	if err := manager.Config.SetUpFromFile(); err != nil {
		panic(err)
	}

	r := router.New(manager)
	// Web UI
	e := echo.New()
	e.HideBanner = true
	pubsub.SystemEvent.Pub(pubsub.System{Time: time.Now(), Type: systemevent.WEBUI_START})

	r.SetUp(e)

	if err := e.Start(":8080"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
