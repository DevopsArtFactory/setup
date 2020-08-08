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
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/GwonsooLee/kubenx/pkg/color"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	configFile = home() + "/.aws/setup"
)

type AssumeList struct {
	Profile		   string `yaml:"profile"`
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

	assumeCreds, err := getAssumeCreds(targetRole.RoleArn, assumeMap.SessionName)
	if err != nil {
		return err
	}

	pbcopy := exec.Command("pbcopy")
	if pbcopy == nil {

	}
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
	aList, err := getTotalAssumeList()
	if err != nil {
		return nil, err
	}

	if len(aList) == 0 {
		return nil, fmt.Errorf("You need to upgrade setup, please use `setup upgrade`")
	}

	profile := viper.GetString("profile")
	for _, as := range aList {
		if as.Profile == profile {
			return &as, nil
		}
	}

	return nil, nil
}

func getTotalAssumeList() ([]AssumeList, error) {
	if !checkFileExists(configFile) {
		return nil, fmt.Errorf("%s does not exist. please use `setup init`", configFile)
	}

	aList := []AssumeList{}

	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	yaml.Unmarshal(file, &aList)

	return aList, nil
}

func getSingleAssumeList() (*AssumeList, error) {
	al := AssumeList{}

	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	yaml.Unmarshal(file, &al)

	return &al, nil
}

func home() string {
	dir, _ := homedir.Dir()
	return dir
}

func AddNewAssumeRole(args []string) error {
	if len(args) > 2 {
		return fmt.Errorf("usage: setup add [key]")
	}

	currentProfile :=  viper.GetString("profile")
	color.Blue.Fprintf(os.Stdout, "current profile: %s", currentProfile)
	aList, err := getTotalAssumeList()
	if err != nil {
		return err
	}

	hasProfile := false
	for _, al := range aList {
		if al.Profile == currentProfile {
			hasProfile = true
		}
	}

	var sessionName string
	if ! hasProfile {
		color.Red.Fprintf(os.Stdout, "register new profile first, %s", currentProfile)
		prompt := &survey.Input{
			Message: "New session name: ",
		}
		survey.AskOne(prompt, &sessionName)

		aList = append(aList, AssumeList{
			Profile:        currentProfile,
			SessionName:    sessionName,
			AssumeRoleList: []AssumeRole{},
		})
	}

	var target string
	if len(args) == 0 {
		prompt := &survey.Input{
			Message: "Key: ",
		}
		survey.AskOne(prompt, &target)

		if len(target) == 0 {
			return fmt.Errorf("you have to input key")
		}
	} else {
		target = args[0]
	}

	var assumeRole string
	prompt := &survey.Input{
		Message: "Role ARN: ",
	}
	survey.AskOne(prompt, &assumeRole)

	if len(assumeRole) == 0 {
		return fmt.Errorf("you have to specify ARN of IAM role")
	}

	for i, al := range aList {
		if al.Profile == currentProfile {
			aList[i].AssumeRoleList = append(al.AssumeRoleList, AssumeRole{
				Key:     target,
				RoleArn: assumeRole,
			})
		}
	}

	if err := SyncFile(aList); err != nil {
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

	aList := []AssumeList{}
	aList = append(aList, AssumeList{
		Profile:        "default",
		SessionName:    sessionName,
	})

	if err := SyncFile(aList); err != nil {
		return err
	}

	return nil
}

func Upgrade() error {
	aList, err := getTotalAssumeList()
	if err != nil {
		return err
	}

	if len(aList) == 0 {
		al, err := getSingleAssumeList()
		if err != nil {
			return err
		}

		if al == nil {
			return fmt.Errorf("you do not have any setup profile, please use `setup init`")
		}

		al.Profile = "default"

		aList = append(aList, *al)

		if err := SyncFile(aList); err != nil {
			return err
		}
		color.Blue.Fprintf(os.Stdout, "successfully upgrade to latest")
	} else {
		color.Blue.Fprintf(os.Stdout, "already latest version setup")
	}

	return nil
}

func SyncFile(aList []AssumeList) error {
	writeData, err := yaml.Marshal(aList)
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
	color.Cyan.Fprintf(out, strings.Join(assumeList, "\n"))

	return nil
}

func EditRole() error {
	out := os.Stdout

	currentProfile :=  viper.GetString("profile")
	color.Blue.Fprintf(out, "current profile: %s", currentProfile)

	aList, err := getTotalAssumeList()
	if err != nil {
		return err
	}

	newA := AssumeList{}
	for _, as := range aList {
		if as.Profile == currentProfile {
			newA = as
		}
	}

	assumeKeys := []string{}
	for _, v := range newA.AssumeRoleList {
		assumeKeys = append(assumeKeys, v.Key)
	}

	var target string
	prompt := &survey.Select{
		Message: "Choose account to edit:",
		Options: assumeKeys,
	}
	survey.AskOne(prompt, &target)

	if len(target) == 0{
		return fmt.Errorf("you have to choose key")
	}

	var assumeRole string
	survey.AskOne(&survey.Input{
		Message: "New role ARN: ",
	}, &assumeRole)

	if len(assumeRole) == 0 {
		return fmt.Errorf("you have to specifiy assume role ARN")
	}

	for i, v := range newA.AssumeRoleList {
		if v.Key == target {
			newA.AssumeRoleList[i].RoleArn = assumeRole
		}
	}

	for i, as := range aList {
		if as.Profile == currentProfile {
			aList[i] = newA
		}
	}

	if err := SyncFile(aList); err != nil {
		return err
	}

	color.Green.Fprintf(out, "Role ARN of %s has been updated", target)

	return nil
}

func DeleteCmd() error {
	out := os.Stdout

	currentProfile :=  viper.GetString("profile")
	color.Blue.Fprintf(out, "current profile: %s", currentProfile)

	aList, err := getTotalAssumeList()
	if err != nil {
		return err
	}

	newA := AssumeList{}
	for _, as := range aList {
		if as.Profile == currentProfile {
			newA = as
		}
	}

	assumeKeys := []string{}
	for _, v := range newA.AssumeRoleList {
		assumeKeys = append(assumeKeys, v.Key)
	}

	var target string
	prompt := &survey.Select{
		Message: "Choose account to delete:",
		Options: assumeKeys,
	}
	survey.AskOne(prompt, &target)

	if target == "" {
		fmt.Errorf("you have to choose key")
	}

	newList := []AssumeRole{}
	for _, v := range newA.AssumeRoleList {
		if v.Key != target {
			newList = append(newList, v)
		}
	}

	newA.AssumeRoleList = newList

	for i, as := range aList {
		if as.Profile == currentProfile {
			aList[i] = newA
		}
	}

	if err := SyncFile(aList); err != nil {
		return err
	}

	color.Green.Fprintf(out, "%s is deleted", target)

	return nil
}

func getAssumeCreds(arn string, session_name string) (*sts.Credentials, error) {
	svc := getSTSSession()
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(arn),
		RoleSessionName: aws.String(session_name),
	}

	result, err := svc.AssumeRole(input)
	if err != nil {
		return nil, err
	}
	return result.Credentials, nil
}

func getSTSSession() *sts.STS {
	ResetAWSEnvironmentVariable()

	awsRegion := viper.GetString("region")
	profile := viper.GetString("profile")

	sess := session.Must(
		session.NewSession(&aws.Config{
			Credentials: credentials.NewCredentials(&credentials.SharedCredentialsProvider{
				Filename: defaults.SharedCredentialsFilename(),
				Profile:  profile,
			}),
		}),
	)
	svc := sts.New(sess, &aws.Config{Region: aws.String(awsRegion)})

	return svc
}

func ResetAWSEnvironmentVariable() {
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_SESSION_TOKEN")
}

func WhoAmI() error {
	stsClient :=  getSTSSessionWithoutReset()

	result, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return err
	}

	fmt.Println(result)

	return nil
}

func getSTSSessionWithoutReset() *sts.STS {
	awsRegion := viper.GetString("region")

	sess := session.Must(session.NewSession())
	svc := sts.New(sess, &aws.Config{Region: aws.String(awsRegion)})

	return svc
}
