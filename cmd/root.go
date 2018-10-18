// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"os/exec"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var Roles = pflag.String("roles", "", "get a list of roles")
var Permission = pflag.String("permission", "", "filters by permission")
var Project = pflag.String("project", "", "projectId to run command on")

//DoNormalGcloud does normal gcloud stuff
func DoNormalGcloud() {
	cmd := exec.Command("gcloud", pflag.Args()...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	fmt.Printf("%s\n", string(out))
}

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:              "gcloudx",
	TraverseChildren: true,
	Short:            "Gcloud Xtended",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

func containsOverride(args []string, currCommand *cobra.Command) bool {
	if len(args) == 0 {
		return true
	}
	top, currargs := args[0], args[1:]
	if currCommand.TraverseChildren != true {
		// log.Printf("found sub %s", currCommand.Use)
		return true
	}
	for _, command := range currCommand.Commands() {
		if command.Use == top {
			return containsOverride(currargs, command)
		}
	}
	return false
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cmdArgs := pflag.Args()
	if containsOverride(cmdArgs, RootCmd) {
		if err := RootCmd.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		DoNormalGcloud()
	}

	// // for _, command := range rootCmd.Commands {
	// for _, sub := range rootCmd.Commands() {
	// 	fmt.Println(sub.Use)
	// }
	// if err := rootCmd.Execute(); err != nil {
	// 	doNormal()
	// 	// fmt.Printl("WHATS IP".err)
	// 	// os.Exit(1)
	// }
}

func init() {
	cobra.OnInitialize(initConfig)
	viper.BindPFlags(pflag.CommandLine)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobragcloud.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cobragcloud" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cobragcloud")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
