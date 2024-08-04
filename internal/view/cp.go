package view

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/rivo/tview"
	"github.com/sanoyo/vislam/internal/ui"
)

// Get cp form content
func (v *view) cpForm() (*tview.Form, *string) {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return nil, nil
	}
	// container containerName
	containerName := *selected.container.Name

	readOnly := ""
	if v.app.ReadOnly {
		readOnly = readOnlyLabel
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
		return f, &title
	}

	// handle form submit
	f.AddButton("Start", func() {
		remote := f.GetFormItemByLabel(remoteLabel).(*tview.Checkbox).IsChecked()
		bucket := f.GetFormItemByLabel(bucketLabel).(*tview.InputField).GetText()
		path := f.GetFormItemByLabel(pathLabel).(*tview.InputField).GetText()
		remotePath := f.GetFormItemByLabel(remotePathLabel).(*tview.InputField).GetText()
		delete := f.GetFormItemByLabel(deleteLabel).(*tview.Checkbox).IsChecked()
		baseDir := "e1s"

		isDir := false
		dirname := ""
		if !remote {
			fileInfo, err := os.Stat(path)
			if err != nil {
				v.app.Notice.Errorf("Path error %v", err)
				v.closeModal()
				return
			}
			isDir = fileInfo.IsDir()
			dirname = filepath.Base(path)
		}

		if remotePath == "" {
			remotePath = "."
		} else if remotePath == "/" {
		} else {
			remotePath = strings.TrimSuffix(remotePath, "/")
		}
		if remote {
			ext := filepath.Ext(remotePath)
			if ext == "" {
				isDir = true
			}
			dirname = filepath.Base(remotePath)
		}

		bin, err := exec.LookPath(awsCli)
		if err != nil {
			v.app.Notice.Warnf("failed to find aws cli binary, error: %v", err)
			v.app.back()
		}

		uploadArgs, downloadArgs := []string{}, []string{}

		// cp objects to s3
		if remote {
			uploadArgs = append(uploadArgs, []string{
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
			}...)
			if isDir {
				uploadArgs = append(uploadArgs, fmt.Sprintf("aws s3 cp %s/ s3://%s/%s/%s --recursive", remotePath, bucket, baseDir, dirname))
			} else {
				uploadArgs = append(uploadArgs, fmt.Sprintf("aws s3 cp %s s3://%s/%s/", remotePath, bucket, baseDir))
			}
		} else {
			uploadArgs = append(uploadArgs, []string{
				"s3",
				"cp",
				path,
			}...)

			if isDir {
				uploadArgs = append(uploadArgs, fmt.Sprintf("s3://%s/%s/%s", bucket, baseDir, dirname))
				uploadArgs = append(uploadArgs, "--recursive")
			} else {
				uploadArgs = append(uploadArgs, fmt.Sprintf("s3://%s/%s/", bucket, baseDir))
			}
		}

		// cp objects from s3
		if remote {
			downloadArgs = append(downloadArgs, []string{
				"s3",
				"cp",
			}...)

			if isDir {
				downloadArgs = append(downloadArgs, fmt.Sprintf("s3://%s/%s", bucket, baseDir))
			} else {
				downloadArgs = append(downloadArgs, fmt.Sprintf("s3://%s/%s/", bucket, baseDir))
			}
			downloadArgs = append(downloadArgs, path)
			downloadArgs = append(downloadArgs, "--recursive")
		} else {
			downloadArgs = append(downloadArgs, []string{
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
			}...)

			if isDir {
				downloadArgs = append(downloadArgs, fmt.Sprintf("aws s3 cp s3://%s/%s/%s %s/%s --recursive", bucket, baseDir, dirname, remotePath, dirname))
			} else {
				downloadArgs = append(downloadArgs, fmt.Sprintf("aws s3 cp s3://%s/%s/ %s/ --recursive", bucket, baseDir, remotePath))
			}
		}

		// delete s3 objects
		deleteArgs := []string{
			"s3",
			"rm",
			fmt.Sprintf("s3://%s/%s", bucket, baseDir),
			"--recursive",
		}

		var stderr bytes.Buffer
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

		v.app.Suspend(func() {
			v.app.isSuspended = true
			slog.Info("exec", "command", bin+" "+strings.Join(uploadArgs, " "))
			uploadCmd := exec.Command(bin, uploadArgs...)
			uploadCmd.Stdin, uploadCmd.Stdout = os.Stdin, os.Stdout
			uploadCmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
			uploadCmd.Stdout.Write([]byte("\nUpload...\n"))
			err = uploadCmd.Run()
			if err != nil {
				return
			}

			time.Sleep(time.Second)

			slog.Info("exec", "command", bin+" "+strings.Join(downloadArgs, " "))
			downloadCmd := exec.Command(bin, downloadArgs...)
			downloadCmd.Stdin, downloadCmd.Stdout = os.Stdin, os.Stdout
			downloadCmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
			downloadCmd.Stdout.Write([]byte("\nDownload...\n"))
			err = downloadCmd.Run()
			if err != nil {
				return
			}

			time.Sleep(time.Second)

			if delete {
				slog.Info("exec", "command", bin+" "+strings.Join(deleteArgs, " "))
				deleteCmd := exec.Command(bin, deleteArgs...)
				deleteCmd.Stdin, deleteCmd.Stdout = os.Stdin, os.Stdout
				deleteCmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
				deleteCmd.Stdout.Write([]byte("\nDelete...\n"))
				err = deleteCmd.Run()
				if err != nil {
					return
				}

				time.Sleep(2 * time.Second)
			}

			signal.Stop(interrupt)
			close(interrupt)
			v.app.isSuspended = false
		})

		if err != nil {
			v.app.Notice.Errorf("Failed to transfer file, %s", stderr.String())
		} else {
			v.app.Notice.Info("File transfer done")
		}

		v.closeModal()
		v.reloadResource(false)
	})
	return f, &title
}
