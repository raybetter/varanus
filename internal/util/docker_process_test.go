package util

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
)

func getHttpHelloWorldOptions() *dockertest.RunOptions {
	return &dockertest.RunOptions{
		Repository: "ghcr.io/infrastructure-as-code/hello-world",
		Tag:        "2.4.0",
		PortBindings: map[docker.Port][]docker.PortBinding{
			"8080/tcp": {{HostIP: "localhost", HostPort: "8000/tcp"}},
		},
	}

}

func TestDockerHelloWorld(t *testing.T) {

	serverGet := func() (string, error) {
		resp, err := http.Get("http://localhost:8000/")
		if err != nil {
			return "", fmt.Errorf("http.Get error: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("http status not okay: %d", resp.StatusCode)
		}
		//read and return the body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read body: %w", err)
		}
		return string(bodyBytes), nil
	}

	operation := DockerContextOperation{
		MaxReadyWait: 60 * time.Second,
		ReadyWait:    1 * time.Second,
		ReadyCallback: func() bool {
			_, err := serverGet()
			if err != nil {
				fmt.Println("error", err)
				return false
			}
			return true
		},
		ExecutionCallback: func() error {
			//this is where we are testing values
			response, err := serverGet()
			assert.Nil(t, err)
			assert.Equal(t, "Hello, World!", response)
			return nil
		},
	}

	options := getHttpHelloWorldOptions()

	err := TheDockerContext.ExecuteWithOptions(
		options,
		&operation,
	)
	assert.Nil(t, err)

}

func TestInvalidDockerContexts(t *testing.T) {
	optionsCallback := func(*docker.HostConfig) {}

	{
		err := TheDockerContext.ExecuteWithOptionsAndOptionsCallback(
			nil,
			optionsCallback,
			&DockerContextOperation{},
		)
		assert.ErrorContains(t, err, "nil options argument")
	}
	{
		err := TheDockerContext.ExecuteWithOptionsAndOptionsCallback(
			&dockertest.RunOptions{},
			nil,
			&DockerContextOperation{},
		)
		assert.ErrorContains(t, err, "nil optionsCallback argument")
	}
	{
		err := TheDockerContext.ExecuteWithOptionsAndOptionsCallback(
			&dockertest.RunOptions{},
			optionsCallback,
			nil,
		)
		assert.ErrorContains(t, err, "nil operation argument")
	}
	{
		err := TheDockerContext.ExecuteWithOptionsAndOptionsCallback(
			&dockertest.RunOptions{},
			optionsCallback,
			&DockerContextOperation{
				ReadyCallback: nil,
				ReadyWait:     time.Duration(1),
			},
		)
		assert.ErrorContains(t, err, "operation's ReadyWait must not be set if ReadyCallback is nil")
	}
	{
		err := TheDockerContext.ExecuteWithOptionsAndOptionsCallback(
			&dockertest.RunOptions{},
			optionsCallback,
			&DockerContextOperation{
				ReadyCallback: nil,
				MaxReadyWait:  time.Duration(1),
			},
		)
		assert.ErrorContains(t, err, "operation's MaxReadyWait must not be set if ReadyCallback is nil")
	}
	{
		err := TheDockerContext.ExecuteWithOptionsAndOptionsCallback(
			&dockertest.RunOptions{},
			optionsCallback,
			&DockerContextOperation{
				ReadyCallback: func() bool { return true },
				ReadyWait:     time.Duration(1),
				MaxReadyWait:  time.Duration(-1),
			},
		)
		assert.ErrorContains(t, err, "operation's MaxReadyWait must be a positive duration")
	}
	{
		err := TheDockerContext.ExecuteWithOptionsAndOptionsCallback(
			&dockertest.RunOptions{},
			optionsCallback,
			&DockerContextOperation{
				ReadyCallback: func() bool { return true },
				ReadyWait:     time.Duration(-1),
				MaxReadyWait:  time.Duration(1),
			},
		)
		assert.ErrorContains(t, err, "operation's ReadyWait must be a positive duration")
	}
}

func TestDockerExecutionFails(t *testing.T) {

	serverGet := func() (string, error) {
		resp, err := http.Get("http://localhost:8000/")
		if err != nil {
			return "", fmt.Errorf("http.Get error: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("http status not okay: %d", resp.StatusCode)
		}
		//read and return the body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read body: %w", err)
		}
		return string(bodyBytes), nil
	}

	operation := DockerContextOperation{
		MaxReadyWait: 60 * time.Second,
		ReadyWait:    1 * time.Second,
		ReadyCallback: func() bool {
			_, err := serverGet()
			if err != nil {
				fmt.Println("error", err)
				return false
			}
			return true
		},
		ExecutionCallback: func() error {
			return fmt.Errorf("intentional failure")
		},
	}

	options := getHttpHelloWorldOptions()

	err := TheDockerContext.ExecuteWithOptions(
		options,
		&operation,
	)
	assert.ErrorContains(t, err, "execution callback error: intentional failure")

}

func TestDockerOperationFailsToStart(t *testing.T) {

	operation := DockerContextOperation{
		ExecutionCallback: func() error {
			assert.FailNow(t, "should never be called")
			return nil
		},
	}

	options := getHttpHelloWorldOptions()
	//zero out the repo name to induce an error
	options.Repository = ""

	err := TheDockerContext.ExecuteWithOptions(
		options,
		&operation,
	)
	assert.ErrorContains(t, err, "could not start Docker resource :2.4.0: no such image")

}

func TestDockerOperationNeverReady(t *testing.T) {

	readyCallbackCount := 0
	callbackElapsedTimes := []time.Duration{}
	lastCallbackTime := time.Time{} //zero value

	readyCallback := func() bool {
		readyCallbackCount += 1
		if !lastCallbackTime.IsZero() {
			callbackElapsedTimes = append(callbackElapsedTimes, time.Since(lastCallbackTime))
		}
		lastCallbackTime = time.Now()

		return false
	}

	operation := DockerContextOperation{
		MaxReadyWait:  2 * time.Second,
		ReadyWait:     500 * time.Millisecond,
		ReadyCallback: readyCallback,
		ExecutionCallback: func() error {
			assert.FailNow(t, "should never be called")
			return nil
		},
	}

	options := getHttpHelloWorldOptions()

	err := TheDockerContext.ExecuteWithOptions(
		options,
		&operation,
	)
	assert.ErrorContains(t, err, "container never reached the ready state after")
	assert.True(t,
		readyCallbackCount >= 4 && readyCallbackCount <= 5,
		"readyCallbackCount not in [4,5] %d", readyCallbackCount)
	for index, elapsed := range callbackElapsedTimes {
		assert.True(t,
			elapsed >= 470*time.Millisecond && elapsed <= 530*time.Millisecond,
			"elapsedCallbackTime not in [470ms,530ms] %s at index %d", elapsed, index)
	}

}
