package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/dixonwille/wmenu/v5"
)
type STSAssumeRoleAPI interface {
	AssumeRole(ctx context.Context,
		params *sts.AssumeRoleInput,
		optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

func TakeRole(c context.Context, api STSAssumeRoleAPI, input *sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
	return api.AssumeRole(c, input)
}

func truncateString(str string, num int) string {
	bnoden := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		bnoden = str[0:num] + ""
	}
	return bnoden
}

func main() {

	// list aws profiles
	awsExecutable, _ := exec.LookPath( "aws" )

	// create the session name from the git username replacing the space with a dash
	sessionNameCmd := "git config user.name | sed 's/[^A-Za-z0-9+=,.@-]//g'"

	sessionNameOut, sessionNameErr := exec.Command("bash", "-c", sessionNameCmd).Output()
	if sessionNameErr != nil {
		fmt.Printf("Failed to execute command: %s", sessionNameCmd)
	}

	sessionNameInitial := string(sessionNameOut)
	// truncate the username to adhere to aws sts requirements
	sessionName := truncateString(sessionNameInitial, 16)
	sessionNameFinal := strings.TrimSpace(sessionName)

	// get a list of the current profiles in ~/.aws/config
  profilesOut, profileErr := exec.Command(awsExecutable, "configure", "list-profiles").Output()
	if profileErr != nil {
		log.Fatal(profileErr)
	}

	// create the initial profile variable
  profile := ""

	// prepare the profiles for the menu
	profiles := string(profilesOut)
  profileListArray := strings.Fields(profiles)
  
	// menu for choosing profile
  menu := wmenu.NewMenu("Which aws profile would you like to use?")
	menu.Action(func (opts []wmenu.Opt) error {profile = opts[0].Text; return nil})
	for i := 0; i < len(profileListArray); i++ {
		if i == 0 {
			menu.Option(profileListArray[i], nil, true, nil)
		} else {
			menu.Option(profileListArray[i], nil, false, nil)
		}
	}

	menuErr := menu.Run()
	if menuErr != nil{
		log.Fatal(menuErr)
	}

	// load the shared aws config (~/.aws/config)
	cfg, cfgErr := config.LoadDefaultConfig(context.TODO(),
    config.WithSharedConfigProfile(profile))
	if cfgErr != nil {
		log.Fatal(cfgErr)
	}

	// create the aws client
	client := iam.NewFromConfig(cfg)

	// list all roles with a path prefix of /aws-reserved/sso.amazonaws.com/
	output, err := client.ListRoles(context.TODO(), &iam.ListRolesInput{
		PathPrefix: aws.String("/aws-reserved/sso.amazonaws.com/"),
	})
	if err != nil{
		log.Fatal(err)
	}

	// choose the first roles arn
	outPutArn := *output.Roles[0].Arn
	roleArn := &outPutArn

	// assume role
	stsClient := sts.NewFromConfig(cfg)

	input := &sts.AssumeRoleInput{
		RoleArn: roleArn,
		RoleSessionName: &sessionNameFinal,
	}
	
	result, err := TakeRole(context.TODO(), stsClient, input)

	if err != nil {
		fmt.Println("AssumeRole Error", err)
		return
  }

	// set credentials in ~/.aws/credentials
	cmdSetAccessKeyId := &exec.Cmd {
		Path: awsExecutable,
		Args: []string{awsExecutable, "configure", "set", "--profile", "default", "aws_access_key_id", *result.Credentials.AccessKeyId},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	if err := cmdSetAccessKeyId.Run(); err != nil {
		fmt.Println("Error:", err)
	}

	cmdSetSecretAccessKey := &exec.Cmd {
		Path: awsExecutable,
		Args: []string{awsExecutable, "configure", "set", "--profile", "default", "aws_secret_access_key", *result.Credentials.SecretAccessKey},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	if err := cmdSetSecretAccessKey.Run(); err != nil {
		fmt.Println("Error:", err)
	}

	cmdSetSessionToken := &exec.Cmd {
		Path: awsExecutable,
		Args: []string{awsExecutable, "configure", "set", "--profile", "default", "aws_session_token", *result.Credentials.SessionToken},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	if err := cmdSetSessionToken.Run(); err != nil {
		fmt.Println("Error:", err)
	}

}

