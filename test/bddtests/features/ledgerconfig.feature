#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@ledgerconfig
Feature: ledger-config

  Background: Setup
    Given the channel "mychannel" is created and all peers have joined
    And fabric-cli network is initialized
    And fabric-cli plugin "../../.build/ledgerconfig" is installed
    And fabric-cli context "org1-context" is defined on channel "mychannel" with org "peerorg1", peers "peer0.org1.example.com,peer1.org1.example.com" and user "User1"
    And fabric-cli context "org2-context" is defined on channel "mychannel" with org "peerorg2", peers "peer0.org2.example.com,peer1.org2.example.com" and user "User1"

  @ledgerconfig_s1
  Scenario: Test the ledgerconfig sub-commands: update, query, and delete
    # Use org1 to save the config
    Given fabric-cli context "org1-context" is used

    Given variable "org1Config" is assigned the JSON value '{"MspID":"Org1MSP","Apps":[{"AppName":"app1","Version":"v1","Components":[{"Name":"comp1","Version":"v1","Config":"org1-app1-comp1-config","Format":"Other"},{"Name":"comp2","Version":"v1","Config":"org1-app1-comp2-config","Format":"Other"}]}]}'
    Given variable "org2Config" is assigned the JSON value '{"MspID":"Org2MSP","Apps":[{"AppName":"app1","Version":"v1","Components":[{"Name":"comp1","Version":"v1","Config":"org2-app1-comp1-config","Format":"Other"},{"Name":"comp2","Version":"v1","Config":"org2-app1-comp2-config","Format":"Other"}]}]}'
    When fabric-cli is executed with args "ledgerconfig update --config ${org1Config} --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --config ${org2Config} --noprompt"
    Then we wait 1 seconds

    # Use org2 to query the config
    Given fabric-cli context "org2-context" is used
    # Query using --criteria
    Given variable "noMatchingCriteria" is assigned the JSON value '{"MspID":"OrgXMSP"}'
    When fabric-cli is executed with args "ledgerconfig query --criteria ${noMatchingCriteria}"
    Then the JSON path "#" of the response has 0 items

    Given variable "org1Criteria" is assigned the JSON value '{"MspID":"Org1MSP","AppName":"app1"}'
    When fabric-cli is executed with args "ledgerconfig query --criteria ${org1Criteria}"
    Then the JSON path "#" of the response has 2 items
    And the JSON path "0.MspID" of the response equals "Org1MSP"
    And the JSON path "0.AppName" of the response equals "app1"
    And the JSON path "1.MspID" of the response equals "Org1MSP"
    And the JSON path "1.AppName" of the response equals "app1"
    And the JSON path "#.ComponentName" of the response contains "comp1"
    And the JSON path "#.ComponentName" of the response contains "comp2"
    And the JSON path "#.Config" of the response contains "org1-app1-comp1-config"
    And the JSON path "#.Config" of the response contains "org1-app1-comp2-config"

    Given variable "org2Comp2Criteria" is assigned the JSON value '{"MspID":"Org2MSP","AppName":"app1","AppVersion":"v1","ComponentName":"comp2","ComponentVersion":"v1"}'
    When fabric-cli is executed with args "ledgerconfig query --criteria ${org2Comp2Criteria}"
    Then the JSON path "#" of the response has 1 items
    And the JSON path "0.Config" of the response equals "org2-app1-comp2-config"

    # Query using flags
    Given variable "noMatchingCriteria" is assigned the JSON value '{"MspID":"OrgXMSP"}'
    When fabric-cli is executed with args "ledgerconfig query --mspid OrgXMSP"
    Then the JSON path "#" of the response has 0 items

    When fabric-cli is executed with args "ledgerconfig query --mspid Org1MSP --appname app1"
    Then the JSON path "#" of the response has 2 items
    And the JSON path "#.ComponentName" of the response contains "comp1"
    And the JSON path "#.ComponentName" of the response contains "comp2"

    When fabric-cli is executed with args "ledgerconfig query --mspid Org2MSP --appname app1 --appver v1 --componentname comp2 --componentver v1"
    Then the JSON path "#" of the response has 1 items
    And the JSON path "0.ComponentName" of the response equals "comp2"

    # Delete using criteria
    When fabric-cli is executed with args "ledgerconfig delete --criteria ${org2Comp2Criteria} --noprompt"
    Then we wait 1 seconds
    When fabric-cli is executed with args "ledgerconfig query --criteria ${org2Comp2Criteria}"
    Then the JSON path "#" of the response has 0 items

    # Delete using flags
    # First make sure the item is there
    When fabric-cli is executed with args "ledgerconfig query --mspid Org1MSP --appname app1 --appver v1 --componentname comp1 --componentver v1"
    Then the JSON path "#" of the response has 1 items
    # Delete the item
    When fabric-cli is executed with args "ledgerconfig delete --mspid Org1MSP --appname app1 --appver v1 --componentname comp1 --componentver v1 --noprompt"
    Then we wait 1 seconds
    # Now make sure the item is not there
    When fabric-cli is executed with args "ledgerconfig query --mspid Org1MSP --appname app1 --appver v1 --componentname comp1 --componentver v1"
    Then the JSON path "#" of the response has 0 items
