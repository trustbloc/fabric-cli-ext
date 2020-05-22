/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package basecmd

import (
	"bufio"
	"fmt"

	"github.com/hyperledger/fabric-cli/cmd/common"
	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/hyperledger/fabric-cli/pkg/fabric"
)

// FactoryProvider creates a new Factory
type FactoryProvider func(config *environment.Config) (fabric.Factory, error)

var defaultFactoryProvider = func(config *environment.Config) (fabric.Factory, error) {
	return fabric.NewFactory(config)
}

// Command is the base for all commands
type Command struct {
	common.Command

	FactoryProvider FactoryProvider
}

// New returns a Command
func New(settings *environment.Settings, p FactoryProvider) *Command {
	c := &Command{}
	if p == nil {
		p = defaultFactoryProvider
	}
	c.FactoryProvider = p
	c.Settings = settings
	return c
}

// Channel returns a new SDK channel
func (c *Command) Channel() (fabric.Channel, error) {
	factory, err := c.FactoryProvider(c.Settings.Config)
	if err != nil {
		return nil, err
	}
	return factory.Channel()
}

// ResMgmt returns a new SDK resource manager
func (c *Command) ResMgmt() (fabric.ResourceManagement, error) {
	factory, err := c.FactoryProvider(c.Settings.Config)
	if err != nil {
		return nil, err
	}

	return factory.ResourceManagement()
}

// Context returns the current context
func (c *Command) Context() *environment.Context {
	return c.Settings.Config.Contexts[c.Settings.Config.CurrentContext]
}

// Fprintln displays the given args to the configured output stream
func (c *Command) Fprintln(arg ...interface{}) error {
	_, err := fmt.Fprintln(c.Settings.Streams.Out, arg...)
	return err
}

// FprintlnOrPanic displays the given args to the configured output stream.
// If an error occurs then this function panics.
func (c *Command) FprintlnOrPanic(arg ...interface{}) {
	if _, err := fmt.Fprintln(c.Settings.Streams.Out, arg...); err != nil {
		panic(err.Error())
	}
}

// Fprint displays the given args to the configured output stream
func (c *Command) Fprint(arg ...interface{}) error {
	_, err := fmt.Fprint(c.Settings.Streams.Out, arg...)
	return err
}

// FprintOrPanic displays the given args to the configured output stream.
// If an error occurs then this function panics.
func (c *Command) FprintOrPanic(arg ...interface{}) {
	if _, err := fmt.Fprint(c.Settings.Streams.Out, arg...); err != nil {
		panic(err.Error())
	}
}

// Prompt waits for the user to enter a string and returns the string
func (c *Command) Prompt() string {
	ackChan := make(chan string)
	go c.readFromTerminal(ackChan)
	ack := <-ackChan
	return ack
}

func (c *Command) readFromTerminal(responsech chan string) {
	reader := bufio.NewReader(c.Settings.Streams.In)
	if response, err := reader.ReadString('\n'); err != nil {
		c.FprintlnOrPanic(fmt.Sprintf("Error reading from terminal: %s", err))
	} else {
		responsech <- response[0:1]
	}
}
