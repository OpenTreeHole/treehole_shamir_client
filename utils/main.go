package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	"io"
	"net/http"
	"os"
	"strconv"
)

var client = http.Client{}

func DecryptAllUser(key *crypto.Key, authUrl string) error {

	var identityName = key.GetEntity().PrimaryIdentity().Name
	fmt.Printf("your uid is %v\n", identityName)

	req, err := http.NewRequest("GET", authUrl+"/api/shamir?identity_name="+identityName, nil)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	_ = rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return errors.New(string(data))
	}

	messages := make([]PGPMessageResponse, 0)
	err = json.Unmarshal(data, &messages)
	if err != nil {
		return err
	}
	fmt.Println("receive messages from server, decrypting...")

	allUser := len(messages)
	shareRequest := UploadSharesRequest{
		IdentityName: identityName,
		Shares:       make([]UserShare, 0, allUser),
	}

	privateKey, err := key.Armor()
	if err != nil {
		return err
	}

	for i, message := range messages {
		shareString, err := helper.DecryptMessageArmored(privateKey, nil, message.PGPMessage)
		if err != nil {
			return err
		}

		share, err := FromString(shareString)
		if err != nil {
			return err
		}
		shareRequest.Shares = append(shareRequest.Shares, UserShare{
			UserID: message.UserID,
			Share:  share,
		})

		fmt.Printf("\ruser_id: %d, (%d / %d)", message.UserID, i, allUser)
	}

	fmt.Println("\nDone!")

	shareData, err := json.Marshal(shareRequest)
	if err != nil {
		return err
	}

	shareFile, err := os.Create("share_data_all.json")
	if err != nil {
		return err
	}

	_, err = shareFile.Write(shareData)
	if err != nil {
		return err
	}

	err = shareFile.Close()
	if err != nil {
		return err
	}

	fmt.Println("share upload request save to share_data_all.json")

	rsp, err = client.Post(authUrl+"/api/shamir/shares", "application/json", bytes.NewBuffer(shareData))
	if err != nil {
		return err
	}

	data, err = io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	_ = rsp.Body.Close()

	if rsp.StatusCode >= 400 {
		return fmt.Errorf("error sending shares: %v", err)
	}

	fmt.Println("shares upload success")

	return nil
}

func DecryptByUserID(key *crypto.Key, userID int, authUrl string) error {
	var identityName = key.GetEntity().PrimaryIdentity().Name
	fmt.Printf("your uid is %v\n", identityName)

	req, err := http.NewRequest(
		"GET",
		authUrl+"/api/shamir/"+strconv.Itoa(userID)+"?identity_name="+identityName,
		nil,
	)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	_ = rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return errors.New(string(data))
	}

	var message PGPMessageResponse
	err = json.Unmarshal(data, &message)
	if err != nil {
		return err
	}
	fmt.Println("receive messages from server, decrypting...")

	privateKey, err := key.Armor()
	if err != nil {
		return err
	}

	shareString, err := helper.DecryptMessageArmored(privateKey, nil, message.PGPMessage)
	if err != nil {
		return err
	}
	_, err = FromString(shareString)
	if err != nil {
		return err
	}

	fmt.Printf("user_id: %v\n%v\n", userID, shareString)

	return nil
}
