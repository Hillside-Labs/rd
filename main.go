package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/ionrock/procs"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/urfave/cli/v2"
)

type tfoutput struct {
	Values struct {
		RootModule struct {
			Resources []struct {
				Values struct {
					Name      string `json:"name"`
					IP        string `json:"ipv4_address"`
					PrivateIP string `json:"ipv4_address_private"`
				} `json:"values"`
			} `json:"resources"`
		} `json:"root_module"`
	} `json:"values"`
}

type Host struct {
	Name      string
	IP        string
	PrivateIP string
}

func (h Host) String() string {
	return fmt.Sprintf("%s:\t%s\t[%s]", h.Name, h.IP, h.PrivateIP)
}

// GetHosts runs a script to get the hosts. In our case it is looking
// at the terraform output in the infra repo.
func GetHosts() ([]Host, error) {
	cmd := exec.Command("terraform", "show", "-json")
	cmd.Dir = "./infra"
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var tf tfoutput
	if err := json.Unmarshal(out, &tf); err != nil {
		return nil, err
	}

	hosts := []Host{}

	for _, value := range tf.Values.RootModule.Resources {
		// Only add a host if it has an address.
		if value.Values.IP != "" && value.Values.PrivateIP != "" {
			hosts = append(hosts, Host{
				Name:      value.Values.Name,
				IP:        value.Values.IP,
				PrivateIP: value.Values.PrivateIP,
			})
		}
	}

	return hosts, nil
}

func GetTargets(name, ip, private string) ([]Host, error) {
	hosts, err := GetHosts()
	if err != nil {
		return nil, err
	}

	targets := []Host{}

	for _, h := range hosts {
		if name != "" && strings.HasPrefix(h.Name, name) {
			targets = append(targets, h)
		}
		if ip != "" && h.IP == ip {
			targets = append(targets, h)
		}
		if private != "" && h.Name == private {
			targets = append(targets, h)
		}

		if name == "" && ip == "" && private == "" {
			targets = append(targets, h)
		}
	}

	return targets, nil
}

func GetTargetsWithFlags(c *cli.Context) ([]Host, error) {
	return GetTargets(
		c.String("name"),
		c.String("ip"),
		c.String("private"),
	)
}

func rdUser() string {
	user := os.Getenv("RD_USER")
	if user == "" {
		user = "root"
	}

	return user
}

func NewSSHCmd(host Host, args ...string) *exec.Cmd {
	user := rdUser()
	conn := fmt.Sprintf("%s@%s", user, host.IP)
	command := exec.Command("ssh", conn)
	command.Args = append(command.Args, args...)

	return command
}

func ExecuteCmd(host Host, args ...string) {
	command := NewSSHCmd(host, args...)
	p := procs.Process{Cmds: []*exec.Cmd{command}}
	p.OutputHandler = func(line string) string {
		fmt.Printf("%s\t| %s\n", host.Name, line)
		return line
	}

	p.ErrHandler = p.OutputHandler

	fmt.Printf("%s\t| Running: %s\n", host.Name, command.String())
	p.Run()
}

func SyncFiles(host Host, src string, recursive bool) {
	user := rdUser()
	dest := fmt.Sprintf("%s@%s:.", user, host.IP)
	command := exec.Command("scp")
	if recursive {
		command.Args = append(command.Args, "-r")
	}
	command.Args = append(command.Args, src, dest)
	p := procs.Process{Cmds: []*exec.Cmd{command}}
	p.OutputHandler = func(line string) string {
		fmt.Printf("%s | %s\n", host.Name, line)
		return line
	}

	fmt.Printf("%s\t| Running: %s\n", host.Name, command.String())
	p.Run()
}

var nameFlag = &cli.StringFlag{
	Name:    "name",
	Usage:   "filter hosts by name",
	Aliases: []string{"n"},
}
var ipFlag = &cli.StringFlag{
	Name:    "ip",
	Usage:   "filter hosts by public ip",
	Aliases: []string{"i"},
}
var privateFlag = &cli.StringFlag{
	Name:    "private",
	Usage:   "filter hosts by public ip",
	Aliases: []string{"p"},
}

func main() {

	app := cli.App{
		Name:  "rd",
		Usage: "rd stands for remote docker",
		Commands: []*cli.Command{
			{
				Name: "bootstrap",
				Flags: []cli.Flag{
					nameFlag, ipFlag, privateFlag},
				Action: func(c *cli.Context) error {
					fname, err := BoostrapScript()
					if err != nil {
						return err
					}

					defer os.Remove(fname)

					targets, err := GetTargetsWithFlags(c)

					if err != nil {
						return err
					}

					for _, t := range targets {
						SyncFiles(t, fname, false)
						ExecuteCmd(t, "chmod", "+x", "bootstrap.sh")

						// TODO: Might need sudo here...
						ExecuteCmd(t, "./bootstrap.sh")
					}

					return nil
				},
			},
			{
				Name: "sync",
				Flags: []cli.Flag{
					nameFlag, ipFlag, privateFlag,
					&cli.BoolFlag{
						Name:    "recursive",
						Aliases: []string{"r"},
						Usage:   "Sync recursively like scp -r",
					},
				},
				Action: func(c *cli.Context) error {
					targets, err := GetTargetsWithFlags(c)

					if err != nil {
						return err
					}

					fn := c.Args().First()
					if fn == "" {
						fn = "docker-compose.yml"
					}

					for _, t := range targets {
						SyncFiles(t, fn, c.Bool("recursive"))
					}
					return nil
				},
			},
			{
				Name:  "run",
				Flags: []cli.Flag{nameFlag, ipFlag, privateFlag},
				Action: func(c *cli.Context) error {
					targets, err := GetTargetsWithFlags(c)
					if err != nil {
						return err
					}
					for _, t := range targets {
						fmt.Println(t)
					}

					for _, t := range targets {
						ExecuteCmd(t, c.Args().Slice()...)
					}

					return nil
				},
			},
			{
				Name: "config",
				Action: func(c *cli.Context) error {

					endpoint := "sfo3.digitaloceanspaces.com"
					useSSL := true
					spacesKey := os.Getenv("AWS_ACCESS_KEY_ID")
					spacesSecret := os.Getenv("AWS_SECRET_ACCESS_KEY")

					bucket := c.Args().First()
					key := c.Args().Get(1)

					if bucket == "" || key == "" {
						return fmt.Errorf("Missing bucket: %s or key: %s", bucket, key)
					}

					minioClient, err := minio.New(endpoint, &minio.Options{
						Creds:  credentials.NewStaticV4(spacesKey, spacesSecret, ""),
						Secure: useSSL,
					})

					if err != nil {
						return err
					}

					obj, err := minioClient.GetObject(context.Background(), bucket, key, minio.GetObjectOptions{})
					if err != nil {
						return err
					}
					defer obj.Close()

					content, err := io.ReadAll(obj)
					if err != nil {
						return err
					}
					fmt.Println(string(content))

					return err

				},
			},
			{
				Name:  "hosts",
				Flags: []cli.Flag{nameFlag, ipFlag, privateFlag},
				Action: func(c *cli.Context) error {
					targets, err := GetTargetsWithFlags(c)
					if err != nil {
						return err
					}

					for _, t := range targets {
						fmt.Println(t.Name, t.IP, t.PrivateIP)
					}

					return nil
				},
			},
		},
	}

	app.Commands = addDockerCmds(app.Commands)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
