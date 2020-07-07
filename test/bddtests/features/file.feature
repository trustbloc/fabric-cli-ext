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

    Given fabric-cli network is initialized
    And fabric-cli plugin "../../.build/ledgerconfig" is installed
    And fabric-cli plugin "../../.build/file" is installed
    And fabric-cli plugin "../../.build/extensions" is installed
    And fabric-cli context "org1-admin-context" is defined on channel "mychannel" with org "peerorg1", peers "peer0.org1.example.com,peer1.org1.example.com" and user "Admin"
    And fabric-cli context "org1-context" is defined on channel "mychannel" with org "peerorg1", peers "peer0.org1.example.com,peer1.org1.example.com" and user "User1"
    And fabric-cli context "org2-context" is defined on channel "mychannel" with org "peerorg2", peers "peer0.org2.example.com,peer1.org2.example.com" and user "User1"

    Given fabric-cli context "org1-admin-context" is used

    Given variable "both-orgs-policy" is assigned the value "AND('Org1MSP.member','Org2MSP.member')"
    And variable "single-org-policy" is assigned the value "OR('Org1MSP.member','Org2MSP.member')"
    And variable "implicit-org-policy" is assigned the value "OR('IMPLICIT-ORG.member')"

    Given variable "sidetreetxn-coll-cfg" is assigned the JSON value '{"name":"dcas","type":"COL_DCAS","policy":"${single-org-policy}","requiredPeerCount":1,"maxPeerCount":2,"timeToLive":"10m"}'
    Given variable "metadata-coll-cfg" is assigned the JSON value '{"name":"meta_data","type":"COL_OFFLEDGER","policy":"${implicit-org-policy}","requiredPeerCount":0,"maxPeerCount":0,"timeToLive":""}'
    Given variable "fileidxdoc-coll-cfg" is assigned the JSON value '{"name":"fileidxdoc","type":"COL_OFFLEDGER","policy":"${implicit-org-policy}","requiredPeerCount":0,"maxPeerCount":0,"timeToLive":""}'
    Given variable "consortium-coll-cfg" is assigned the JSON value '{"name":"consortium","type":"COL_DCAS","policy":"${single-org-policy}","requiredPeerCount":1,"maxPeerCount":2,"timeToLive":""}'

    Then fabric-cli is executed with args "extensions instantiatecc configscc v1 --policy ${both-orgs-policy}" ignoring error ".*chaincode with name '.*' already exists.*"
    And fabric-cli is executed with args "extensions instantiatecc sidetreetxn v1 --policy ${both-orgs-policy} --collections-config [${sidetreetxn-coll-cfg}]" ignoring error ".*chaincode with name '.*' already exists.*"
    And fabric-cli is executed with args "extensions instantiatecc document v1 --policy ${single-org-policy} --collections-config [${fileidxdoc-coll-cfg},${metadata-coll-cfg}]" ignoring error ".*chaincode with name '.*' already exists.*"
    And fabric-cli is executed with args "extensions instantiatecc file v1 --policy ${single-org-policy} --collections-config [${consortium-coll-cfg}]" ignoring error ".*chaincode with name '.*' already exists.*"

    Then we wait 10 seconds

    Given fabric-cli context "org1-context" is used
    Then fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/consortium-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/org1-config.json --noprompt"
    And fabric-cli is executed with args "ledgerconfig update --configfile ./fixtures/config/fabric/org2-config.json --noprompt"

    Given variable "token_content_w" is assigned the value "TOKEN_CONTENT_W"
    And variable "token_fileidx_r" is assigned the value "TOKEN_FILEIDX_R"
    And variable "token_fileidx_w" is assigned the value "TOKEN_FILEIDX_W"

    Then we wait 10 seconds

  @sidetree_file_s1
  Scenario: Test the file command
    # Create a file index document
    When fabric-cli is executed with args "file createidx --path /content --url http://localhost:48326/file --recoverykeyfile ./fixtures/testdata/keys/recover/public.key --updatekeyfile ./fixtures/testdata/keys/update/public.key --authtoken ${token_fileidx_w} --noprompt"
    And the JSON path "id" of the response is saved to variable "fileIdxID"

    # Update the file handler configuration for the '/content' path with the ID of the file index document
    Then fabric-cli is executed with args "ledgerconfig fileidxupdate --msp Org1MSP --peers peer0.org1.example.com;peer1.org1.example.com --path /content --idxid ${fileIdxID} --noprompt"
    And fabric-cli is executed with args "ledgerconfig fileidxupdate --msp Org2MSP --peers peer0.org2.example.com;peer1.org2.example.com --path /content --idxid ${fileIdxID} --noprompt"

    Then we wait 10 seconds

    Given the authorization bearer token for "GET" requests to path "/file" is set to "${token_fileidx_r}"
    When an HTTP GET is sent to "http://localhost:48326/file/${fileIdxID}"
    Then the JSON path "didDocument.id" of the response equals "${fileIdxID}"

    # Upload a couple of files and add them to the file index document
    # NOTE: Use an explicit --contentauthtoken to test the case where the auth token for /file and /content are different. Otherwise,
    # if they're the same, we don't need to specify --contentauthtoken.
    When fabric-cli is executed with args "file upload --url http://localhost:48326/content --files ./fixtures/testdata/v1/arrays.schema.json;./fixtures/testdata/v1/geographical-location.schema.json --idxurl http://localhost:48326/file/${fileIdxID} --nextupdatekeyfile ./fixtures/testdata/keys/update/public.key --signingkeyfile ./fixtures/testdata/keys/update/private.key --authtoken ${token_fileidx_w} --contentauthtoken ${token_content_w} --noprompt"
    Then the JSON path "#" of the response has 2 items
    And the JSON path "0.Name" of the response equals "arrays.schema.json"
    And the JSON path "0.ContentType" of the response equals "application/json"
    And the JSON path "1.Name" of the response equals "geographical-location.schema.json"
    And the JSON path "1.ContentType" of the response equals "application/json"

    Then we wait 10 seconds

    When an HTTP GET is sent to "http://localhost:48326/content/arrays.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/arrays.schema.json"

    When an HTTP GET is sent to "http://localhost:48526/content/arrays.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/arrays.schema.json"

    When an HTTP GET is sent to "http://localhost:48626/content/arrays.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/arrays.schema.json"

    When an HTTP GET is sent to "http://localhost:48326/content/geographical-location.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/geographical-location.schema.json"

    When an HTTP GET is sent to "http://localhost:48426/content/geographical-location.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/geographical-location.schema.json"

    When an HTTP GET is sent to "http://localhost:48526/content/geographical-location.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/geographical-location.schema.json"

    When an HTTP GET is sent to "http://localhost:48626/content/geographical-location.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/geographical-location.schema.json"

    # Upload more files and add them to the file index document. Note that arrays.schema.json is updated to v2
    # NOTE: Don't use an explicit --contentauthtoken to test the case where the auth token for /file and /content are the same
    When fabric-cli is executed with args "file upload --url http://localhost:48326/content --files ./fixtures/testdata/v1/person.schema.json;./fixtures/testdata/v1/raised-hand.png;./fixtures/testdata/v1/text1.txt;./fixtures/testdata/v2/arrays.schema.json --idxurl http://localhost:48326/file/${fileIdxID} --nextupdatekeyfile ./fixtures/testdata/keys/update/public.key --signingkeyfile ./fixtures/testdata/keys/update/private.key --authtoken ${token_fileidx_w} --noprompt"
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

    When an HTTP GET is sent to "http://localhost:48326/content/person.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/person.schema.json"
    When an HTTP GET is sent to "http://localhost:48426/content/raised-hand.png"
    When an HTTP GET is sent to "http://localhost:48526/content/text1.txt"

    # Ensure that the original files can still be accessed
    When an HTTP GET is sent to "http://localhost:48526/content/geographical-location.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/geographical-location.schema.json"

    When an HTTP GET is sent to "http://localhost:48626/content/arrays.schema.json"
    Then the JSON path "$id" of the response equals "https://example.com/arrays.schema.v2.json"
