package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	tbPath := "/etc/default/mnm"
	if runtime.GOOS == "darwin" {
		tbPath = "/etc/defaults/mnm"
	}
	tb, _ := os.ReadFile(tbPath)
	token := strings.TrimSpace(string(tb))
	app := &cli.App{
		Name:    "mnm",
		Usage:   "monitor & notifier to messenger",
		Version: "0.0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "api",
				Value: "https://mnm.sh",
				Usage: "The webhook api",
			},
			&cli.StringFlag{
				Name:  "token",
				Value: fmt.Sprintf("%s", token),
				Usage: fmt.Sprintf("The webhook token (%s)", tbPath),
			},
		},
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:   "run",
				Usage:  "Run a command",
				Action: action,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

func action(c *cli.Context) error {
	startAt := time.Now()
	api := c.String("api")
	token := c.String("token")
	prog := c.Args().First()

	parts := strings.Split(prog, " ")
	name, args := parts[0], parts[1:]
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("invalid command to run %s", prog)
	}
	cmd := exec.Command(name, args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, p := range []io.ReadCloser{stdout, stderr} {
		wg.Add(1)
		go func(pipe io.ReadCloser) {
			defer wg.Done()
			io.Copy(os.Stdout, pipe)
		}(p)
	}
	wg.Wait()

	result, err := "OK", cmd.Wait()
	if err != nil {
		result = err.Error()
	}

	return notify(api, token, prog, result, startAt)
}

func notify(api, token, run, result string, startAt time.Time) error {
	endpoint := api + "/in/" + token
	runtime := time.Now().Sub(startAt).String()
	info := fmt.Sprintf("RUN: %s\r\nRESULT: %s\r\nRUNTIME: %s", run, result, runtime)
	body, _ := json.Marshal(map[string]string{
		"category": "PLAIN_TEXT",
		"data":     base64.URLEncoding.EncodeToString([]byte(info)),
	})
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
