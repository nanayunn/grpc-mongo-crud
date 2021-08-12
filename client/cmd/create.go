/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	userpb "example.com/grpc-mongo-crud/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a New User",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create called")

		// requestCtx, _ := context.WithTimeout(context.Background(), 10*time.Second)

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			status.Errorf(
				codes.InvalidArgument,
				fmt.Sprintf("Invalid Flag values, please check the values : %v", err),
			)
		}
		age, err := cmd.Flags().GetString("age")
		if err != nil {
			status.Errorf(
				codes.InvalidArgument,
				fmt.Sprintf("Invalid Flag values, please check the values : %v", err),
			)
		}
		userid, err := cmd.Flags().GetString("userid")
		if err != nil {
			status.Errorf(
				codes.InvalidArgument,
				fmt.Sprintf("Invalid Flag values, please check the values : %v", err),
			)
		}
		user := &userpb.User{
			Name:   name,
			Age:    age,
			Userid: userid,
		}

		fmt.Println(user)
		res, err := client.CreateUser(
			requestCtx,
			&userpb.CreateUserReq{
				User: user,
			},
		)
		if err != nil {
			fmt.Printf("err : %v\n", err)
		}
		fmt.Printf("\nUser Created\n")
		fmt.Printf("%v\n", res)

	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringP("name", "n", "", "User name")
	createCmd.Flags().StringP("age", "a", "", "User age")
	createCmd.Flags().StringP("userid", "u", "", "User ID")
	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("age")
	createCmd.MarkFlagRequired("userid")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
