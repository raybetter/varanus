package util

import (
	"fmt"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// Interface

type DockerContext interface {
	ExecuteWithOptionsAndOptionsCallback(
		options *dockertest.RunOptions,
		optionsCallback func(*docker.HostConfig),
		operation *DockerContextOperation,
	) error
	ExecuteWithOptions(
		options *dockertest.RunOptions,
		operation *DockerContextOperation,
	) error
}

type DockerContextOperation struct {
	//when this time has expired, we give up on the container becoming ready and abort
	MaxReadyWait time.Duration
	// How long between ready calls
	ReadyWait time.Duration
	//this will be called at ReadyWait intervals until it returns true; abort after
	ReadyCallback func() bool
	//this is executed once ready returns true
	ExecutionCallback func() error
}

// Singleton
var TheDockerContext DockerContext

func init() {
	TheDockerContext = new(dockerContextImpl)
}

// Implementation

type dockerContextImpl struct {
	pool *dockertest.Pool
}

// Init sets up the dockerContextImpl.  Calls to it are idempotent.
func (dc *dockerContextImpl) init() error {
	if dc.pool == nil {
		pool, err := dockertest.NewPool("")

		if err != nil {
			//no coverage because requires inducing docker errors, beyond scope
			return fmt.Errorf("could not construct pool: %w", err)
		}

		dc.pool = pool
	}

	return nil
}

func (dc *dockerContextImpl) ExecuteWithOptions(
	options *dockertest.RunOptions,
	operation *DockerContextOperation,
) error {
	return dc.ExecuteWithOptionsAndOptionsCallback(
		options,
		func(*docker.HostConfig) {},
		operation,
	)
}

func (dc *dockerContextImpl) ExecuteWithOptionsAndOptionsCallback(
	options *dockertest.RunOptions,
	optionsCallback func(*docker.HostConfig),
	operation *DockerContextOperation,
) error {
	if options == nil {
		return fmt.Errorf("nil options argument")
	}
	if optionsCallback == nil {
		return fmt.Errorf("nil optionsCallback argument")
	}
	if operation == nil {
		return fmt.Errorf("nil operation argument")
	}
	if operation.ReadyCallback != nil {
		//if there is a ready callback, then the other ready options need to be set
		if operation.MaxReadyWait <= 0 {
			return fmt.Errorf("operation's MaxReadyWait must be a positive duration")
		}
		if operation.ReadyWait <= 0 {
			return fmt.Errorf("operation's ReadyWait must be a positive duration")
		}
	} else {
		//an error to set the waits without providing a ready callback
		if operation.MaxReadyWait != 0 {
			return fmt.Errorf("operation's MaxReadyWait must not be set if ReadyCallback is nil")
		}
		if operation.ReadyWait != 0 {
			return fmt.Errorf("operation's ReadyWait must not be set if ReadyCallback is nil")
		}

	}

	dc.init()

	// uses pool to try to connect to Docker
	if err := dc.pool.Client.Ping(); err != nil {
		//no coverage because requires inducing docker errors, beyond scope
		return fmt.Errorf("could not connect to Docker: %w", err)
	}

	cleanupCallback := func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := dc.pool.RunWithOptions(options, optionsCallback, cleanupCallback)
	if err != nil {
		return fmt.Errorf("could not start Docker resource %s:%s: %w", options.Repository, options.Tag, err)
	}

	//set up a deferral to clean up the docker container at the end of execution
	defer func() {
		err = dc.pool.Purge(resource)
		if err != nil {
			//no coverage because requires inducing docker errors, beyond scope
			err = fmt.Errorf("could not purge resource for with %s:%s: %w", options.Repository, options.Tag, err)
		}
	}()

	//do the ready sequence if we have a ready callback
	if operation.ReadyCallback != nil {
		readyStart := time.Now()

		var readyResult bool = false
		for !readyResult {
			readyResult = operation.ReadyCallback()

			readyElapsed := time.Since(readyStart)
			if readyElapsed > operation.MaxReadyWait {
				return fmt.Errorf("container never reached the ready state after %s", readyElapsed)
			}

			if !readyResult {
				time.Sleep(operation.ReadyWait)
			}
		}
	}

	//execute the callback
	if err := operation.ExecutionCallback(); err != nil {
		return fmt.Errorf("execution callback error: %w", err)
	}

	//cleanup is in the deferred function

	return nil
}
