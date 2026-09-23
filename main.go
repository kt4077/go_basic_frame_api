package main

import "server_api/cmd"

func main() {
	if err := cmd.RootCmd().Execute(); err != nil {
		panic(err)
	}
}
