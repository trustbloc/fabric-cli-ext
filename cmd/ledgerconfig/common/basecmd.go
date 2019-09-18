/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

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

// BaseCommand is the base for all commands
type BaseCommand struct {
	common.Command

	FactoryProvider FactoryProvider
}

// NewBaseCmd returns a BaseCommand
func NewBaseCmd(settings *environment.Settings, p FactoryProvider) *BaseCommand {
	c := &BaseCommand{}
	if p == nil {
		p = defaultFactoryProvider
	}
	c.FactoryProvider = p
	c.Settings = settings
	return c
}

// Channel returns a new SDK channel
func (c *BaseCommand) Channel() (fabric.Channel, error) {
	factory, err := c.FactoryProvider(c.Settings.Config)
	if err != nil {
		return nil, err
	}
	return factory.Channel()
}

// Context returns the current context
func (c *BaseCommand) Context() *environment.Context {
	return c.Settings.Config.Contexts[c.Settings.Config.CurrentContext]
}

// Fprintln displays the given args to the configured output stream
func (c *BaseCommand) Fprintln(arg ...interface{}) error {
	_, err := fmt.Fprintln(c.Settings.Streams.Out, arg...)
	return err
}

// FprintlnOrPanic displays the given args to the configured output stream.
// If an error occurs then this function panics.
func (c *BaseCommand) FprintlnOrPanic(arg ...interface{}) {
	if _, err := fmt.Fprintln(c.Settings.Streams.Out, arg...); err != nil {
		panic(err.Error())
	}
}

// Prompt waits for the user to enter a string and returns the string
func (c *BaseCommand) Prompt() string {
	ackChan := make(chan string)
	go c.readFromTerminal(ackChan)
	ack := <-ackChan
	return ack
}

func (c *BaseCommand) readFromTerminal(responsech chan string) {
	reader := bufio.NewReader(c.Settings.Streams.In)
	if response, err := reader.ReadString('\n'); err != nil {
		c.FprintlnOrPanic(fmt.Sprintf("Error reading from terminal: %s", err))
	} else {
		responsech <- response[0:1]
	}
}
