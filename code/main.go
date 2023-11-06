package main

import (
	"code/client"
	"code/config_parser"
	"code/daemon"
	"code/pomodoro"
	"fmt"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Pomodoro timer",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var pomodoro_states []pomodoro.PomodoroStates
		var err error

		if len(args) == 0 {
			pomodoro_states, err = config_parser.GetPomodoroStates()
			fmt.Println("Starting the Pomodoro timer with default settings:")
		} else {
			arg := args[0]
			pomodoro_states, err = config_parser.Parse(arg)
			fmt.Println("Starting the Pomodoro timer with custom settings:")
		}
		if err != nil {
			fmt.Println(err)
			return
		}

		d := daemon.NewDaemon(pomodoro_states)
		d.Start()

		for _, state := range pomodoro_states {
			fmt.Printf("%v: %v min\n", state.State, state.Time)
		}

	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Pomodoro timer",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := client.NewPomodoroIPCClient()
		if err != nil {
			fmt.Println(err)
			return
		}
		response, err := client.SendCommand("stop")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(response)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the Pomodoro timer status",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := client.NewPomodoroIPCClient()
		if err != nil {
			fmt.Println(err)
			return
		}
		response, err := client.SendQuery("status")
		if err != nil {
			fmt.Println(err)
			return
		}
		var output string

		for _, status := range response.States {
			output = fmt.Sprintf("State: %v, Time: %v\n", status.State, status.Time)
			if status.State == response.Current {
				output = fmt.Sprintf("%v <-- Current", output)
			}
			fmt.Print(output)
		}
	},
}

func main() {
	var rootCmd = &cobra.Command{Use: "pomodoro"}
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
