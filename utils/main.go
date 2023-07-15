package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var client = http.Client{}

var Key *crypto.Key

var KeyRing *crypto.KeyRing

func DecryptAllUser(authUrl string) error {

	var identityName = Key.GetEntity().PrimaryIdentity().Name
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

	messageChan := make(chan PGPMessageResponse)
	resultChan := make(chan UserShareError)
	done, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start PGP decrypt goroutines
	for i := 0; i < runtime.NumCPU(); i++ {
		go decryptMessageTask(messageChan, resultChan, done)
	}

	// start message sending goroutine
	go func() {
		for i, message := range messages {
			messageChan <- message
			fmt.Printf("\ruser_id: %d, (%d / %d)", message.UserID, i+1, allUser)
		}
	}()

	// receive shares
	for i := 0; i < allUser; i++ {
		result := <-resultChan
		if result.error != nil {
			return err
		} else {
			shareRequest.Shares = append(shareRequest.Shares, result.UserShare)
		}
	}

	fmt.Println("\nDone!")

	shareData, err := SaveShareData(shareRequest, identityName)
	if err != nil {
		return err
	}

	err = UploadShares(shareData, authUrl)
	if err != nil {
		return err
	}

	return nil
}

func SaveShareData(shareRequest UploadSharesRequest, identityName string) (shareData []byte, err error) {
	shareData, _ = json.Marshal(shareRequest)

	filename := fmt.Sprintf("share_data_%s.json", strings.Split(identityName, " ")[0])

	err = os.WriteFile(filename, shareData, 0666)
	if err != nil {
		return nil, err
	}

	fmt.Printf("share upload request save to %s\n", filename)

	return shareData, nil
}

func UploadShares(shareData []byte, authUrl string) error {

	rsp, err := client.Post(authUrl+"/api/shamir/shares", "application/json", bytes.NewBuffer(shareData))
	if err != nil {
		return err
	}

	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	_ = rsp.Body.Close()

	if rsp.StatusCode >= 400 {
		return fmt.Errorf("error sending shares: %v", string(data))
	}

	fmt.Println("shares upload success")
	return nil
}

func DecryptByUserID(userID int, authUrl string) error {
	var identityName = Key.GetEntity().PrimaryIdentity().Name
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

	shareString, err := decryptMessage(message)

	fmt.Printf("user_id: %v\n%s\n", userID, shareString)

	return nil
}

func decryptMessageTask(
	messageChan <-chan PGPMessageResponse,
	resultChan chan<- UserShareError,
	ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case message := <-messageChan:
			shareString, err := decryptMessage(message)
			if err != nil {
				resultChan <- UserShareError{
					error: err,
				}
			} else {
				share, err := FromString(shareString)
				if err != nil {
					resultChan <- UserShareError{
						error: err,
					}
				}
				resultChan <- UserShareError{
					UserShare: UserShare{
						UserID: message.UserID,
						Share:  share,
					},
					error: nil,
				}
			}
		}
	}
}

func decryptMessage(message PGPMessageResponse) (string, error) {
	pgpMessage, err := crypto.NewPGPMessageFromArmored(message.PGPMessage)
	if err != nil {
		return "", err
	}

	rawMessage, err := KeyRing.Decrypt(pgpMessage, nil, 0)
	return rawMessage.GetString(), err
}
