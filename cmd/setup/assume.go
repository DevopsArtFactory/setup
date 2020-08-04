/*
Copyright © 2020 DevopsArtFactory gwonsoo.lee@gmail.com

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
package setup

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/GwonsooLee/kubenx/pkg/aws"
	"github.com/GwonsooLee/kubenx/pkg/color"
)

var (
	configFile = home() + "/.aws/setup"
)

type AssumeList struct {
	SessionName    string       `yaml:"session_name"`
	AssumeRoleList []AssumeRole `yaml:"assume_role_list"`
}

type AssumeRole struct {
	Key     string `yaml:"key"`
	RoleArn string `yaml:"role_arn"`
}

func Assume(out io.Writer, args []string) error {
	var err error

	assumeMap, err := getAssumeList()
	if err != nil {
		return err
	}

	var target string
	if len(args) == 0 {
		assumeList := []string{}
		for _, v := range assumeMap.AssumeRoleList {
			assumeList = append(assumeList, v.Key)
		}

		prompt := &survey.Select{
			Message: "Choose account:",
			Options: assumeList,
		}
		survey.AskOne(prompt, &target)

		if target == "" {
			color.Red.Fprintln(out, fmt.Errorf("changing Context has been canceled"))
			return nil
		}
	} else {
		target = args[0]
	}

	var targetRole AssumeRole
	for _, v := range assumeMap.AssumeRoleList {
		if v.Key == target {
			targetRole = v
			break
		}
	}

	if targetRole.Key == "" {
		return fmt.Errorf("wrong command. please check `setup help")
	}

	assumeCreds := aws.AssumeRole(targetRole.RoleArn, assumeMap.SessionName)

	pbcopy := exec.Command("pbcopy")
	in, _ := pbcopy.StdinPipe()

	if err := pbcopy.Start(); err != nil {
		return err
	}

	if _, err := in.Write([]byte(fmt.Sprintf("export AWS_ACCESS_KEY_ID=%s\n", *assumeCreds.AccessKeyId))); err != nil {
		return err
	}

	if _, err := in.Write([]byte(fmt.Sprintf("export AWS_SECRET_ACCESS_KEY=%s\n", *assumeCreds.SecretAccessKey))); err != nil {
		return err
	}

	if _, err := in.Write([]byte(fmt.Sprintf("export AWS_SESSION_TOKEN=%s\n", *assumeCreds.SessionToken))); err != nil {
		return err
	}

	if err := in.Close(); err != nil {
		return err
	}

	err = pbcopy.Wait()
	if err != nil {
		color.Red.Fprintln(out, err.Error())
		return err
	}

	color.Blue.Fprintln(out, "Assume Credentials copied to clipboard, please paste it.")
	return nil
}

func getAssumeList() (*AssumeList, error) {
	if !checkFileExists(configFile) {
		return nil, fmt.Errorf("%s does not exist. please use `setup init`", configFile)
	}

	a := AssumeList{}

	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	yaml.Unmarshal(file, &a)

	return &a, nil
}

func home() string {
	dir, _ := homedir.Dir()
	return dir
}

func AddNewAssumeRole(args []string) error {
	if len(args) > 2 {
		return fmt.Errorf("usage: setup add [key]")
	}

	var target string
	if len(args) == 0 {
		prompt := &survey.Input{
			Message: "Key: ",
		}
		survey.AskOne(prompt, &target)
	} else {
		target = args[0]
	}

	var assumeRole string
	prompt := &survey.Input{
		Message: "Role ARN: ",
	}
	survey.AskOne(prompt, &assumeRole)

	if assumeRole == "" {
		return fmt.Errorf("you have to specify ARN of IAM role")
	}

	al, err := getAssumeList()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	al.AssumeRoleList = append(al.AssumeRoleList, AssumeRole{
		Key:     target,
		RoleArn: assumeRole,
	})

	if err := SyncFile(*al); err != nil {
		return err
	}

	return nil
}

func checkFileExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func Setup() error {
	if checkFileExists(configFile) {
		return fmt.Errorf("you already have %s file", configFile)
	}

	var sessionName string
	prompt := &survey.Input{
		Message: "Your session name: ",
	}
	survey.AskOne(prompt, &sessionName)

	al := AssumeList{}
	al.SessionName = sessionName

	if err := SyncFile(al); err != nil {
		return err
	}

	return nil
}

func SyncFile(al AssumeList) error {
	writeData, err := yaml.Marshal(al)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(configFile, writeData, 0644); err != nil {
		return err
	}

	return nil
}

func ListRole() error {
	al, err := getAssumeList()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	assumeList := []string{}
	for _, v := range al.AssumeRoleList {
		assumeList = append(assumeList, v.Key)
	}

	out := os.Stdout
	color.Green.Fprintf(out, "[current role list]")
	color.Cyan.Fprintf(out, strings.Join(assumeList,"\n"))

	return nil
}

func EditRole() error {
	out := os.Stdout
	al, err := getAssumeList()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	assumeList := []string{}
	for _, v := range al.AssumeRoleList {
		assumeList = append(assumeList, v.Key)
	}

	var target string
	prompt := &survey.Select{
		Message: "Choose account to edit:",
		Options: assumeList,
	}
	survey.AskOne(prompt, &target)

	if target == "" {
		fmt.Errorf("you have to choose key")
	}

	var assumeRole string
	survey.AskOne(&survey.Input{
		Message: "New role ARN: ",
	}, &assumeRole)

	if assumeRole == "" {
		fmt.Errorf("you have to specifiy assume role ARN")
	}

	for i, v := range al.AssumeRoleList {
		if v.Key == target {
			al.AssumeRoleList[i].RoleArn = assumeRole
		}
	}

	if err := SyncFile(*al); err != nil {
		return err
	}

	color.Green.Fprintf(out, "Role ARN of %s has been updated", target)

	return nil
}

func DeleteCmd() error {
	out := os.Stdout
	al, err := getAssumeList()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	assumeList := []string{}
	for _, v := range al.AssumeRoleList {
		assumeList = append(assumeList, v.Key)
	}

	var target string
	prompt := &survey.Select{
		Message: "Choose account to delete:",
		Options: assumeList,
	}
	survey.AskOne(prompt, &target)

	if target == "" {
		fmt.Errorf("you have to choose key")
	}

	newList := []AssumeRole{}
	for _, v := range al.AssumeRoleList {
		if v.Key != target {
			newList = append(newList, v)
		}
	}

	al.AssumeRoleList = newList

	if err := SyncFile(*al); err != nil {
		return err
	}

	color.Green.Fprintf(out, "%s is deleted", target)

	return nil
}
