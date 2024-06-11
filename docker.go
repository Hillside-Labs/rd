package main

import "github.com/urfave/cli/v2"

var dockerStatusCmd = []string{"docker", "compose", "ps"}

func addDockerCmds(cmd []*cli.Command) []*cli.Command {
	cmd = append(cmd, []*cli.Command{
		{
			Name:  "start",
			Flags: []cli.Flag{nameFlag, ipFlag, privateFlag},
			Action: func(c *cli.Context) error {
				targets, err := GetTargetsWithFlags(c)
				if err != nil {
					return err
				}

				args := []string{
					"docker", "compose", "up", "-d",
				}
				if c.Args().First() != "" {
					args = append(args, c.Args().First())
				}

				for _, t := range targets {
					ExecuteCmd(t, args...)
					ExecuteCmd(t, dockerStatusCmd...)
				}

				return nil

			},
		},
		{
			Name:  "pull",
			Flags: []cli.Flag{nameFlag, ipFlag, privateFlag},
			Action: func(c *cli.Context) error {
				targets, err := GetTargetsWithFlags(c)
				if err != nil {
					return err
				}

				args := []string{
					"docker", "compose", "pull",
				}
				if c.Args().First() != "" {
					args = append(args, c.Args().First())
				}

				for _, t := range targets {
					ExecuteCmd(t, args...)
				}

				return nil
			},
		},
		{
			Name:  "create",
			Flags: []cli.Flag{nameFlag, ipFlag, privateFlag},
			Action: func(c *cli.Context) error {
				targets, err := GetTargetsWithFlags(c)
				if err != nil {
					return err
				}

				args := []string{
					"docker", "compose", "pull",
				}
				if c.Args().First() != "" {
					args = append(args, c.Args().First())
				}

				for _, t := range targets {
					ExecuteCmd(t, args...)
				}

				return nil
			},
		},

		{
			Name:  "update",
			Flags: []cli.Flag{nameFlag, ipFlag, privateFlag},
			Action: func(c *cli.Context) error {
				targets, err := GetTargetsWithFlags(c)
				if err != nil {
					return err
				}

				args := []string{
					"docker", "compose", "pull",
				}
				if c.Args().First() != "" {
					args = append(args, c.Args().First())
				}

				for _, t := range targets {
					ExecuteCmd(t, args...)
				}

				for _, t := range targets {
					ExecuteCmd(t, "docker", "ps")
				}

				return nil
			},
		},

		{
			Name:  "restart",
			Flags: []cli.Flag{nameFlag, ipFlag, privateFlag},
			Action: func(c *cli.Context) error {
				targets, err := GetTargetsWithFlags(c)
				if err != nil {
					return err
				}

				args := []string{
					"docker", "compose", "restart",
				}
				if c.Args().First() != "" {
					args = append(args, c.Args().First())
				}

				for _, t := range targets {
					ExecuteCmd(t, args...)
				}

				for _, t := range targets {
					ExecuteCmd(t, "docker", "ps")
				}

				return nil
			},
		},

		{
			Name:  "reboot",
			Flags: []cli.Flag{nameFlag, ipFlag, privateFlag},
			Action: func(c *cli.Context) error {
				targets, err := GetTargetsWithFlags(c)
				if err != nil {
					return err
				}

				svc := c.Args().First()

				stop := []string{
					"docker", "compose", "stop",
				}
				if svc != "" {
					stop = append(stop, svc)
				}

				for _, t := range targets {
					ExecuteCmd(t, stop...)
				}

				rm := []string{
					"docker", "compose", "stop",
				}
				if svc != "" {
					rm = append(rm, svc)
				}

				for _, t := range targets {
					ExecuteCmd(t, rm...)
				}

				up := []string{
					"docker", "compose", "up", "-d",
				}
				if svc != "" {
					up = append(up, svc)
				}

				for _, t := range targets {
					ExecuteCmd(t, up...)
				}

				for _, t := range targets {
					ExecuteCmd(t, "docker", "ps")
				}

				return nil
			},
		},
		{
			Name:  "logs",
			Flags: []cli.Flag{nameFlag, ipFlag, privateFlag},
			Action: func(c *cli.Context) error {
				targets, err := GetTargetsWithFlags(c)
				if err != nil {
					return err
				}
				args := []string{"docker", "compose", "logs", "-f"}
				svc := c.Args().First()
				if svc != "" {
					args = append(args, svc)
				}

				for _, t := range targets {
					ExecuteCmd(t, args...)
				}

				return nil
			},
		},
	}...)

	return cmd
}
