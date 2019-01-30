// Copyright 2019 The Trias Co., LTD. All rights reserved.
// Use of this source code is governed by a GNU General Public License v3.0
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	interval = time.Second * 2
	uri      = "http://localhost:46657/broadcast_tx_commit"
)

type Tx struct {
	Name   string
	Number int
}

var mapCmd = map[string]func(){}

func tmInit() {
	if tmExist() {
		cmd := exec.Command("pkill", "tendermint")
		_, err := cmd.Output()
		if err != nil {
			fmt.Printf("kill tendermint error: %v", err)
			return
		}
	}
	tmDir := getTmDir()
	err := os.RemoveAll(tmDir)
	if err != nil {
		panic(fmt.Errorf("remove tm folder error: %v", err))
	}
	err = os.Mkdir(tmDir, 0755)
	if err != nil {
		panic(fmt.Errorf("make tm folder error: %v", err))
	}

	cmd := exec.Command("tendermint", "init")
	_, err = cmd.Output()
	if err != nil {
		fmt.Printf("initialize tendermint error: %v", err)
		return
	}

	fmt.Println("initialize tendermint success!")
}

func tmExist() bool {
	cmd := exec.Command("pgrep", "tendermint")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return output != nil
}

func getTmDir() string{
	//usr, err := user.Current()
	//if err != nil {
	//	panic(fmt.Errorf("get user error: %v", err))
	//}
	//
	//tmDir := path.Join(usr.HomeDir, ".tendermint")

	tmDir := "/trias/.tendermint"
	return tmDir
}

func tmRun() {
	fmt.Println("tendermint runing")
	cmd := exec.Command("tendermint", "node", "--proxy_app", "nilapp", "--consensus.create_empty_blocks=false")
	stderr, _ := cmd.StdoutPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stderr)
	//scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}
	cmd.Wait()
}

func tx() {
	i := 1
	for {
		s := uuid.New().String()
		s = strings.Repeat(s, 20000)
		data := fmt.Sprintf("tx=\"%s\"", s)
		_, err := http.Post(uri, "application/x-www-form-urlencoded", strings.NewReader(data))
		if err != nil {
			panic(err)
		}

		time.Sleep(interval)
		i++
	}
}

func help() {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)
	defer w.Flush()

	fmt.Println("tmassi is a tendermint assit")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println()
	writeUsage(w, "init", "-i", "tendermint init")
	writeUsage(w, "run", "-r", "node --proxy_app nilapp --consensus.create_empty_blocks=false")
	writeUsage(w, "tx", "-t", "start a loop for sending transactions to tendermint")
}

func writeUsage(w *tabwriter.Writer, cmd, abbr, desp string) {
	fmt.Fprintf(w, "\t%s\t%s\t%s\t\n", cmd, abbr, desp)
}

func main() {
	mapCmd["init"] = tmInit
	mapCmd["-i"] = tmInit

	mapCmd["run"] = tmRun
	mapCmd["-r"] = tmRun

	mapCmd["tx"] = tx
	mapCmd["-t"] = tx

	if len(os.Args) < 2 {
		help()
		return
	}

	cmd := os.Args[1]
	fn, ok := mapCmd[cmd]
	if !ok {
		fmt.Printf("Command \"%s\" not exists\nn", cmd)
		help()
	}

	fn()
}
