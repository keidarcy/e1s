package view

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/keidarcy/e1s/internal/ui"
	"github.com/rivo/tview"
)

// Get cp form content
func (v *view) cpForm() (*tview.Form, string) {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return nil, ""
	}
	// container containerName
	containerName := *selected.container.Name

	readOnly := ""
	if v.app.ReadOnly {
		readOnly = readonlyLabel
	}

	title := " File transfer though S3 bucket [purple::b]" + containerName + readOnly

	f := ui.StyledForm(title)
	remoteLabel := "Remote to local"
	bucketLabel := "Bucket"
	pathLabel := "Local path"
	remotePathLabel := "Remote container path"
	deleteLabel := "Remove S3 objects after transfer"

	f.AddCheckbox(remoteLabel, false, nil)
	f.AddInputField(bucketLabel, "", 50, nil, nil)
	f.AddInputField(pathLabel, "", 50, nil, nil)
	f.AddInputField(remotePathLabel, "", 50, nil, nil)
	f.AddCheckbox(deleteLabel, true, nil)

	// handle form close
	f.AddButton("Cancel", func() {
		v.closeModal()
	})

	// readonly mode has no submit button
	if v.app.ReadOnly {
		return f, title
	}

	// handle form submit
	f.AddButton("Start", func() {
		remote := f.GetFormItemByLabel(remoteLabel).(*tview.Checkbox).IsChecked()
		bucket := f.GetFormItemByLabel(bucketLabel).(*tview.InputField).GetText()
		path := f.GetFormItemByLabel(pathLabel).(*tview.InputField).GetText()
		remotePath := f.GetFormItemByLabel(remotePathLabel).(*tview.InputField).GetText()
		delete := f.GetFormItemByLabel(deleteLabel).(*tview.Checkbox).IsChecked()

		logger.Info(remote, bucket, path, remotePath, delete)

		fileInfo, err := os.Stat(path)
		if err != nil {
			v.app.Notice.Errorf("Path error %v", err)
			v.closeModal()
			return
		}
		isDir := fileInfo.IsDir()
		dirname := filepath.Base(path)

		remotePath = strings.TrimSuffix(remotePath, "/")

		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

		bin, err := exec.LookPath(awsCli)
		if err != nil {
			logger.Warnf("Failed to find aws cli binary, error: %v", err)
			v.app.Notice.Warnf("Failed to find aws cli binary, error: %v", err)
			v.app.back()
		}
		baseDir := "hello-world"
		uploadArgs := []string{
			"s3",
			"cp",
			path,
		}

		if isDir {
			uploadArgs = append(uploadArgs, fmt.Sprintf("s3://%s/%s/%s", bucket, baseDir, dirname))
			uploadArgs = append(uploadArgs, "--recursive")
		} else {
			uploadArgs = append(uploadArgs, fmt.Sprintf("s3://%s/%s/", bucket, baseDir))
		}

		downloadArgs := []string{
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

		if isDir {
			downloadArgs = append(downloadArgs, fmt.Sprintf("aws s3 cp s3://%s/%s/%s %s/%s --recursive", bucket, baseDir, dirname, remotePath, dirname))
		} else {
			downloadArgs = append(downloadArgs, fmt.Sprintf("aws s3 cp s3://%s/%s/ %s/ --recursive", bucket, baseDir, remotePath))
		}

		deleteArgs := []string{
			"s3",
			"rm",
			fmt.Sprintf("s3://%s/%s", bucket, baseDir),
			"--recursive",
		}

		v.app.Suspend(func() {
			logger.Infof("Exec: `%s %s`", awsCli, strings.Join(uploadArgs, " "))
			uploadCmd := exec.Command(bin, uploadArgs...)
			uploadCmd.Stdin, uploadCmd.Stdout, uploadCmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			uploadCmd.Run()

			time.Sleep(time.Second)

			logger.Infof("Exec: `%s %s`", awsCli, strings.Join(downloadArgs, " "))
			downloadCmd := exec.Command(bin, downloadArgs...)
			downloadCmd.Stdin, downloadCmd.Stdout, downloadCmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			downloadCmd.Run()

			time.Sleep(time.Second)

			if delete {
				logger.Infof("Exec: `%s %s`", awsCli, strings.Join(deleteArgs, " "))
				deleteCmd := exec.Command(bin, deleteArgs...)
				deleteCmd.Stdin, deleteCmd.Stdout, deleteCmd.Stderr = os.Stdin, os.Stdout, os.Stderr
				deleteCmd.Run()

				time.Sleep(time.Second)
			}

			signal.Stop(interrupt)
			close(interrupt)
		})

		v.app.Notice.Info("File transfer done")

		v.closeModal()
		v.reloadResource(false)
	})
	return f, title
}
