#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@sidetree-file
Feature: Upload files to DCAS which are backed by a Sidetree file index document

  Background: Setup
    Given the channel "mychannel" is created and all peers have joined

    Then we wait 10 seconds

    Given DCAS collection config "dcas-mychannel" is defined for collection "dcas" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=1, maxPeerCount=2, and timeToLive=60m
    Given DCAS collection config "docs-mychannel" is defined for collection "docs" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=1, maxPeerCount=2, and timeToLive=60m
    Given off-ledger collection config "meta_data_coll" is defined for collection "meta_data" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=0, maxPeerCount=0, and timeToLive=60m
    Given DCAS collection config "consortium-files-coll" is defined for collection "consortium" as policy="OR('Org1MSP.member','Org2MSP.member')", requiredPeerCount=1, maxPeerCount=2, and timeToLive=60m

    Then "test" chaincode "configscc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy ""
    And "system" chaincode "configscc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy ""
    And "system" chaincode "sidetreetxn_cc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "AND('Org1MSP.member','Org2MSP.member')" with collection policy "dcas-mychannel"
    And "system" chaincode "document_cc" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" with collection policy "docs-mychannel,meta_data_coll"
    And "system" chaincode "files" is instantiated from path "in-process" on the "mychannel" channel with args "" with endorsement policy "OR('Org1MSP.member','Org2MSP.member')" with collection policy "consortium-files-coll"

    Then we wait 10 seconds

    Given fabric-cli network is initialized
    And fabric-cli plugin "../../.build/ledgerconfig" is installed
    And fabric-cli plugin "../../.build/file" is installed
    And fabric-cli context "org1-context" is defined on channel "mychannel" with org "peerorg1", peers "peer0.org1.example.com,peer1.org1.example.com" and user "User1"
    And fabric-cli context "org2-context" is defined on channel "mychannel" with org "peerorg2", peers "peer0.org2.example.com,peer1.org2.example.com" and user "User1"

    Given fabric-cli context "org1-context" is used
    Then fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/consortium-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/org1-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/org2-config.json --noprompt"

    Then we wait 10 seconds

  @sidetree_file_s1
  Scenario: Test the file command
    # Create a file index document
    When fabric-cli is executed with args "file createidx --path /content --url http://localhost:48326/file --recoverypwd pwd1 --nextpwd pwd1 --noprompt"
    And the JSON path "id" of the response is saved to variable "fileIdxID"

    # Update the file handler configuration for the '/content' path with the ID of the file index document
    Then fabric-cli is executed with args "ledgerconfig fileidxupdate --msp Org1MSP --peers peer0.org1.example.com;peer1.org1.example.com --path /content --idxid ${fileIdxID} --noprompt"
    And fabric-cli is executed with args "ledgerconfig fileidxupdate --msp Org2MSP --peers peer0.org2.example.com;peer1.org2.example.com --path /content --idxid ${fileIdxID} --noprompt"

    Then we wait 10 seconds

    When an HTTP request is sent to "http://localhost:48326/file/${fileIdxID}"
    Then the JSON path "id" of the response equals "${fileIdxID}"

    # Upload a couple of files and add them to the file index document
    When fabric-cli is executed with args "file upload --url http://localhost:48326/content --files ./fixtures/testdata/v1/arrays.schema.json;./fixtures/testdata/v1/geographical-location.schema.json --idxurl http://localhost:48326/file/${fileIdxID} --pwd pwd1 --nextpwd pwd2 --noprompt"
    Then the JSON path "#" of the response has 2 items
    And the JSON path "0.Name" of the response equals "arrays.schema.json"
    And the JSON path "0.ContentType" of the response equals "application/json"
    And the JSON path "1.Name" of the response equals "geographical-location.schema.json"
    And the JSON path "1.ContentType" of the response equals "application/json"

    Then we wait 10 seconds

    When an HTTP request is sent to "http://localhost:48326/content/arrays.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/arrays.schema.json"

    When an HTTP request is sent to "http://localhost:48426/content/arrays.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/arrays.schema.json"

    When an HTTP request is sent to "http://localhost:48526/content/arrays.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/arrays.schema.json"

    When an HTTP request is sent to "http://localhost:48626/content/arrays.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/arrays.schema.json"

    When an HTTP request is sent to "http://localhost:48326/content/geographical-location.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/geographical-location.schema.json"

    When an HTTP request is sent to "http://localhost:48426/content/geographical-location.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/geographical-location.schema.json"

    When an HTTP request is sent to "http://localhost:48526/content/geographical-location.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/geographical-location.schema.json"

    When an HTTP request is sent to "http://localhost:48626/content/geographical-location.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/geographical-location.schema.json"

    # Upload more files and add them to the file index document. Note that arrays.schema.json is updated to v2
    When fabric-cli is executed with args "file upload --url http://localhost:48326/content --files ./fixtures/testdata/v1/person.schema.json;./fixtures/testdata/v1/raised-hand.png;./fixtures/testdata/v1/text1.txt;./fixtures/testdata/v2/arrays.schema.json --idxurl http://localhost:48326/file/${fileIdxID} --pwd pwd2 --nextpwd pwd3 --noprompt"
    Then the JSON path "#" of the response has 4 items
    And the JSON path "0.Name" of the response equals "person.schema.json"
    And the JSON path "0.ContentType" of the response equals "application/json"
    And the JSON path "1.Name" of the response equals "raised-hand.png"
    And the JSON path "1.ContentType" of the response equals "image/png"
    And the JSON path "2.Name" of the response equals "text1.txt"
    And the JSON path "2.ContentType" of the response equals "text/plain; charset=utf-8"
    And the JSON path "3.Name" of the response equals "arrays.schema.json"
    And the JSON path "3.ContentType" of the response equals "application/json"

    Then we wait 10 seconds

    When an HTTP request is sent to "http://localhost:48326/content/person.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/person.schema.json"
    When an HTTP request is sent to "http://localhost:48426/content/raised-hand.png"
    When an HTTP request is sent to "http://localhost:48526/content/text1.txt"

    # Ensure that the original files can still be accessed
    When an HTTP request is sent to "http://localhost:48526/content/geographical-location.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/geographical-location.schema.json"

    When an HTTP request is sent to "http://localhost:48626/content/arrays.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/arrays.schema.v2.json"