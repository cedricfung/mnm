package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/process"
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
		Version: "0.2.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "api",
				Value: "https://mnm.sh",
				Usage: "The webhook api",
			},
			&cli.StringFlag{
				Name:  "token",
				Value: token,
				Usage: fmt.Sprintf("The webhook token (%s)", tbPath),
			},
		},
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "Run a command, e.g. mnm r 'wget https://some.large/file.zip'",
				Action:  action,
			},
			{
				Name:    "monitor",
				Aliases: []string{"m"},
				Usage:   "Monitor a PID, e.g. mnm m 1314",
				Action:  monitor,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

func monitor(c *cli.Context) error {
	startAt := time.Now()
	api := c.String("api")
	token := c.String("token")
	pid, err := strconv.Atoi(c.Args().First())
	if err != nil || pid <= 0 {
		return fmt.Errorf("invalid PID %s", c.Args().First())
	}

	info := fmt.Sprintf("🟢 MONITOR: %d\r\n🧭 START: %s", pid, startAt)
	err = notify(api, token, info)
	if err != nil {
		return err
	}

	var result error
	for {
		running, err := process.PidExistsWithContext(context.Background(), int32(pid))
		fmt.Printf("PID: %d RUNNING: %t ERROR: %v\n", pid, running, err)
		if err != nil || !running {
			result = err
			break
		}
		time.Sleep(time.Second * 5)
	}

	runtime := time.Since(startAt).String()
	info = fmt.Sprintf("🟢🟢🟢🟢🟢🟢🟢\r\n🚀 MONITOR: %d\r\n🌞 RESULT: %v\r\n🧭 RUNTIME: %s",
		pid, result, runtime)
	if result != nil {
		info = fmt.Sprintf("🔴🔴🔴🔴🔴🔴🔴\r\n🚀 RUN: %d\r\n🚨 RESULT: %v\r\n🧭 RUNTIME: %s",
			pid, result, runtime)
	}
	return notify(api, token, info)
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
	info := fmt.Sprintf("🚀 RUN: %s\r\n🌞 PID: %d\r\n🧭 START: %s", prog, cmd.Process.Pid, startAt)
	err = notify(api, token, info)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, p := range []io.ReadCloser{stdout, stderr} {
		wg.Add(1)
		go func(pipe io.ReadCloser) {
			defer wg.Done()
			_, err := io.Copy(os.Stdout, pipe)
			if err != nil {
				panic(err)
			}
		}(p)
	}
	wg.Wait()

	result, err := "OK", cmd.Wait()
	if err != nil {
		result = err.Error()
	}

	runtime := time.Since(startAt).String()
	info = fmt.Sprintf("🟢🟢🟢🟢🟢🟢🟢\r\n🚀 RUN: %s\r\n🌞 RESULT: %s\r\n🧭 RUNTIME: %s",
		prog, result, runtime)
	if result != "OK" {
		info = fmt.Sprintf("🔴🔴🔴🔴🔴🔴🔴\r\n🚀 RUN: %s\r\n🚨 RESULT: %s\r\n🧭 RUNTIME: %s",
			prog, result, runtime)
	}
	return notify(api, token, info)
}

func notify(api, token, info string) error {
	endpoint := api + "/in/" + token
	body, _ := json.Marshal(map[string]string{
		"category": "PLAIN_TEXT",
		"data":     base64.RawURLEncoding.EncodeToString([]byte(info)),
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
