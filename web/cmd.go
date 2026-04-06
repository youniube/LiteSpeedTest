package web

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/xxf098/lite-proxy/web/render"
)

func readConfig(configPath string) (*ProfileTestOptions, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	options := &ProfileTestOptions{}
	if err = json.Unmarshal(data, options); err != nil {
		return nil, err
	}
	if options.Concurrency < 1 {
		options.Concurrency = 1
	}
	if options.Language == "" {
		options.Language = "en"
	}
	if options.Theme == "" {
		options.Theme = "rainbow"
	}
	if options.Timeout < 8 {
		options.Timeout = 8
	}
	options.Timeout = options.Timeout * time.Second
	return options, nil
}

func TestFromCMD(subscription string, configPath *string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	options := ProfileTestOptions{
		GroupName:       "Default",
		SpeedTestMode:   "all",
		PingMethod:      "googleping",
		SortMethod:      "rspeed",
		Concurrency:     2,
		TestMode:        2,
		Subscription:    subscription,
		Language:        "en",
		FontSize:        24,
		Theme:           "rainbow",
		Timeout:         15 * time.Second,
		GeneratePicMode: PIC_PATH,
		OutputMode:      PIC_PATH,
	}
	if configPath != nil {
		if opt, err := readConfig(*configPath); err == nil {
			options = *opt
			if options.GeneratePicMode != 0 {
				options.OutputMode = options.GeneratePicMode
			}
		}
	}
	if len(subscription) > 0 && subscription != options.Subscription {
		if _, err := url.Parse(subscription); err == nil {
			options.Subscription = subscription
		} else if _, err := os.Stat(subscription); err == nil {
			options.Subscription = subscription
		}
	}
	if jsonOpt, err := json.Marshal(options); err == nil {
		log.Printf("json options: %s\n", string(jsonOpt))
	}
	_, err := TestContext(ctx, options, &OutputMessageWriter{})
	return err
}

func TestContext(ctx context.Context, options ProfileTestOptions, w MessageWriter) (render.Nodes, error) {
	links, err := ParseLinks(options.Subscription)
	if err != nil {
		return nil, err
	}
	p := ProfileTest{
		Writer:      w,
		MessageType: 1,
		Links:       links,
		Options:     &options,
	}
	return p.testAll(ctx)
}

func TestAsyncContext(ctx context.Context, options ProfileTestOptions) (chan render.Node, []string, error) {
	links, err := ParseLinks(options.Subscription)
	if err != nil {
		return nil, nil, err
	}
	p := ProfileTest{
		Writer:      nil,
		MessageType: ALLTEST,
		Links:       links,
		Options:     &options,
	}
	nodeChan, err := p.TestAll(ctx, nil)
	return nodeChan, links, err
}
