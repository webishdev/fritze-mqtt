package main

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/webishdev/fritze-mqtt/fritzbox"
	"github.com/webishdev/fritze-mqtt/internal"
	"github.com/webishdev/fritze-mqtt/log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
)

var Version = "development"
var GitHash = "none"

var showVersion = false
var listOnly = false
var baseUrl string
var username string
var password string
var brokerHost string
var brokerPort int
var mqttTopic string

var sigs chan os.Signal
var controllerTeardown chan byte
var mqttTeardown chan byte

var versionMessage = fmt.Sprintf("Fritze MQTT (Version: %s, Hash: %s)", Version, GitHash)

var rootCmd = &cobra.Command{
	Use:   "fritze-mqtt",
	Short: versionMessage,
	Long:  versionMessage,
	Run: func(cmd *cobra.Command, args []string) {
		err := do()
		if err != nil {
			printError(err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	},
}

func do() error {
	fmt.Println(versionMessage)
	if showVersion {
		return nil
	}

	fmt.Println()

	if username == "" {
		username = os.Getenv("USERNAME")
	}

	if password == "" {
		password = os.Getenv("PASSWORD")
	}

	if username == "" || password == "" {
		ex, err := os.Executable()
		if err != nil {
			return err
		}
		exPath := filepath.Dir(ex)
		envFileAtEx := filepath.Join(exPath, ".env")

		envFileAtExExists := true
		if _, err := os.Stat(envFileAtEx); errors.Is(err, os.ErrNotExist) {
			envFileAtExExists = false
		}

		err = godotenv.Load()
		if err != nil && envFileAtExExists {
			_ = godotenv.Load(envFileAtEx)
		}
	}

	if username == "" {
		username = os.Getenv("USERNAME")
	}

	if password == "" {
		password = os.Getenv("PASSWORD")
	}

	if username == "" || password == "" {
		fmt.Println("username and password required")
		os.Exit(1)
	}

	client := fritzbox.NewFritzClient(baseUrl)

	if listOnly {
		err := internal.ListDevices(client, username, password)
		if err != nil {
			printError(err)
		}
		return nil
	}

	sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	controllerTeardown = make(chan byte, 1)
	mqttTeardown = make(chan byte, 1)

	go func() {
		<-sigs
		log.Info("Received SIGINT/SIGTERM")
		mqttTeardown <- 1
		controllerTeardown <- 1
	}()

	var wg sync.WaitGroup

	go func() {
		defer wg.Done()
		err := internal.StartController(controllerTeardown, client, username, password)
		if err != nil {
			fmt.Println(err)
		}
	}()
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := internal.StartMQTT(mqttTeardown, brokerHost, brokerPort, mqttTopic)
		if err != nil {
			fmt.Println(err)
		}
	}()
	wg.Add(1)

	wg.Wait()

	return nil
}

func printError(current error) {
	_, err := fmt.Fprintln(os.Stderr, "\n", current.Error())
	if err != nil {
		panic(err)
	}
}

func Execute() {
	rootCmd.Flags().SortFlags = false
	rootCmd.Flags().BoolVar(&showVersion, "version", false, "displays the current version")
	rootCmd.Flags().BoolVar(&listOnly, "list", false, "list devices and exit")
	rootCmd.Flags().StringVar(&baseUrl, "base-url", "https://192.168.178.1", "base url of the device")
	rootCmd.Flags().StringVarP(&username, "username", "u", "", "username with smart home rights (env: USERNAME)")
	rootCmd.Flags().StringVarP(&password, "password", "p", "", "password of the user (env: PASSWORD)")
	rootCmd.Flags().StringVar(&brokerHost, "broker-host", "localhost", "hostname of the MQTT broker (env: MQTT_BROKER_HOST)")
	rootCmd.Flags().IntVar(&brokerPort, "broker-port", 1883, "port of the MQTT broker (env: MQTT_BROKER_PORT)")
	rootCmd.Flags().StringVar(&mqttTopic, "topic", "test", "MQTT topic to subscribe (env: MQTT_BROKER_TOPIC)")
	if executeError := rootCmd.Execute(); executeError != nil {
		os.Exit(1)
	}
}

func main() {
	Execute()
}
