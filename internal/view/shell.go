package view

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/keidarcy/e1s/internal/ui"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

// Exec shell to selected container(like ssh)
func (v *view) execShell() {
	args, containerName, err := v.preValidateExec()
	if err != nil {
		v.app.Notice.Warnf("Exec command validation failed: %v", err)
		v.app.back()
		return
	}

	// catch ctrl+C & SIGTERM
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	v.app.Suspend(func() {
		v.app.isSuspended = true
		bin, _ := exec.LookPath(awsCli)
		cmdArgs := append(*args, v.app.Option.Shell)
		slog.Info("exec", "command", bin+" "+strings.Join(cmdArgs, " "))

		cmd := exec.Command(bin, cmdArgs...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		// ignore the stderr from container
		_, err = cmd.Stdout.Write([]byte(fmt.Sprintf(execBannerFmt, *v.app.cluster.ClusterName, *v.app.service.ServiceName, utils.ArnToName(v.app.task.TaskArn), containerName)))
		err = cmd.Run()

		// return signal
		signal.Stop(interrupt)
		close(interrupt)
		v.app.isSuspended = false
	})
}

// Get exec command form content
func (v *view) execCommandForm() (*tview.Form, *string) {
	args, containerName, err := v.preValidateExec()
	if err != nil {
		v.app.Notice.Warnf("Exec command validation failed: %v", err)
		v.app.back()
		return nil, nil
	}

	if containerName == "" {
		return nil, nil
	}

	title := fmt.Sprintf(" Execute command on [purple::b] %s [-:-:-] container?", containerName)
	f := ui.StyledForm(title)
	execLabel := "Execute command"
	f.AddInputField(execLabel, v.app.Option.Shell, 50, nil, nil)

	// handle form close
	f.AddButton("Cancel", func() {
		v.closeModal()
	})

	// handle form submit
	f.AddButton("Execute", func() {
		execCmd := f.GetFormItemByLabel(execLabel).(*tview.InputField).GetText()

		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
		v.app.Suspend(func() {
			v.app.isSuspended = true
			bin, _ := exec.LookPath(awsCli)
			cmdArgs := append(*args, execCmd)
			slog.Info("exec", "command", bin+" "+strings.Join(cmdArgs, " "))

			cmd := exec.Command(bin, cmdArgs...)
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			_, err = cmd.Stdout.Write([]byte(fmt.Sprintf(execBannerFmt, *v.app.cluster.ClusterName, *v.app.service.ServiceName, utils.ArnToName(v.app.task.TaskArn), containerName)))
			time.Sleep(1 * time.Second)

			cmd.Stdout.Write([]byte(fmt.Sprintf("\nExecute: \"%s\"\n", execCmd)))
			err = cmd.Run()

			cmd.Stdout.Write([]byte("\nDone...\n"))
			time.Sleep(3 * time.Second)

			signal.Stop(interrupt)
			close(interrupt)

			v.closeModal()
			v.reloadResource(false)
			v.app.isSuspended = false
		})
	})
	return f, &title
}

func (v *view) preValidateExec() (*[]string, string, error) {
	if v.app.kind != ContainerKind {
		return nil, "", nil
	}

	if v.app.ReadOnly {
		return nil, "", fmt.Errorf("no ecs exec permission in read only e1s mode")
	}

	_, err := exec.LookPath(awsCli)
	if err != nil {
		return nil, "", fmt.Errorf("failed to find %s path, please check %s", awsCli, "https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html")
	}

	selected, err := v.getCurrentSelection()
	if err != nil {
		return nil, "", fmt.Errorf("failed to handleSelected, err: %v", err)
	}

	if selected.container == nil {
		return nil, "", fmt.Errorf("empty pointer selected.container")
	}

	containerName := *selected.container.Name

	_, err = exec.LookPath(smpCi)
	if err != nil {
		return nil, "", fmt.Errorf("failed to find %s path, please check %s", smpCi, "https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html")
	}

	args := []string{
		"ecs",
		"execute-command",
		"--cluster",
		*v.app.cluster.ClusterName,
		"--task",
		*v.app.task.TaskArn,
		"--container",
		containerName,
		"--interactive",
		"--command",
	}

	return &args, containerName, nil
}

// Start session for instance
// aws ssm start-session --target ${instance_id}
func (v *view) instanceStartSession() {
	if v.app.kind != InstanceKind && v.app.kind != TaskKind {
		v.app.Notice.Warn("Invalid kind type to start session")
		return
	}

	if v.app.ReadOnly {
		v.app.Notice.Warn("No permission to start session in read only mode")
		return
	}

	_, err := exec.LookPath(awsCli)
	if err != nil {
		v.app.Notice.Warnf("failed to find %s path, please check %s", awsCli, "https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html")
		return
	}

	selected, err := v.getCurrentSelection()
	if err != nil {
		v.app.Notice.Warnf("failed to handleSelected, err: %v", err)
		return
	}

	instanceId := ""
	if v.app.kind == InstanceKind {
		if selected.instance == nil {
			v.app.Notice.Warn("Not a valid instance")
			return
		}
		instanceId = *selected.instance.Ec2InstanceId
	} else if v.app.kind == TaskKind {
		if v.app.task.ContainerInstanceArn == nil {
			v.app.Notice.Warn("Not a valid task with container instance")
			return
		}
		instanceId, err = v.app.Store.GetTaskInstanceId(v.app.cluster.ClusterName, v.app.task.ContainerInstanceArn)
		if err != nil {
			v.app.Notice.Warnf("failed to get task instance id, err: %v", err)
			return
		}
	}

	if instanceId == "" {
		v.app.Notice.Warn("Not a valid instance")
		return
	}

	_, err = exec.LookPath(smpCi)
	if err != nil {
		v.app.Notice.Warnf("failed to find %s path, please check %s", smpCi, "https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html")
		return
	}

	args := []string{
		"ssm",
		"start-session",
		"--target",
		instanceId,
	}

	// catch ctrl+C & SIGTERM
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	v.app.Suspend(func() {
		v.app.isSuspended = true
		bin, _ := exec.LookPath(awsCli)
		slog.Info("exec", "command", bin+" "+strings.Join(args, " "))

		cmd := exec.Command(bin, args...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		_, err = cmd.Stdout.Write([]byte(fmt.Sprintf(instanceBannerFmt, *v.app.cluster.ClusterName, instanceId)))
		err = cmd.Run()

		// return signal
		signal.Stop(interrupt)
		close(interrupt)
		v.app.isSuspended = false
	})
}
