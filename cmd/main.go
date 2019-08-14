package main

import (
	"fmt"
	"github.com/urfave/cli"
	"mget"
	"net/url"
	"os"
	"strings"
)

func main() {
	app := cli.NewApp()
	app.Name = "mget"
	app.Usage = "a multi-thread downloader"
	app.Version = "1.0.0"
	app.Flags = []cli.Flag{
		cli.UintFlag{
			Name:  "n",
			Value: 4,
			Usage: "number of thread(goroutine actually) to use, 4 default, 99 max",
		},
		cli.StringFlag{
			Name:  "f",
			Usage: "specify the path to save",
		},
	}

	app.Action = func(ctx *cli.Context) error {
		if len(ctx.Args()) != 1 {
			fmt.Printf("\033[31mUsage: mget [-n num_of_thread] [-f path_to_save] url_of_file\033[0m\n")
			return nil
		}

		n := ctx.Uint("n")
		if n == 0 {
			n = 4
		} else if n > 100 {
			n = 99
		}

		_url, err := url.Parse(ctx.Args().First())
		if err != nil || _url.Host == "" {
			fmt.Printf("\033[31mit's a invalid url: %s\033[0m\n", ctx.Args().First())
			return nil
		}

		path := ctx.String("f")
		if path == "" {
			p := strings.Split(_url.Path, "?")[0]
			for i := len(p) - 1; i >= 0; i-- {
				if p[i] == '/' {
					path = p[i+1:]
					break
				}
			}

			if path == "" {
				path = "default"
				fmt.Println("\033[33mthe default path to save file is './default'\033[0m\n")
			}
		}

		if err := mget.Download(ctx.Args().First(), path, int(n)); err != nil {
			fmt.Printf("\033[31mfailed to download the file: %s\033[0m\n", err.Error())
		}
		return nil
	}

	app.Run(os.Args)
}
