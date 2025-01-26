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
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var client = http.Client{}

var Key *crypto.Key

var KeyRing *crypto.KeyRing

func DecryptAllUser(authUrl string, token string) error {

	var identityName = Key.GetEntity().PrimaryIdentity().Name
	fmt.Printf("your identity_name is %v\n", identityName)

	req, err := http.NewRequest("GET", authUrl+"/api/shamir?identity_name="+url.QueryEscape(identityName), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	// fmt.Println(req.URL.String())

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
		fmt.Println(rsp.StatusCode)
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

	err = UploadShares(shareData, authUrl, token)
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

func UploadShares(shareData []byte, authUrl string, token string) error {
	req, err := http.NewRequest("POST", authUrl+"/api/shamir/shares", bytes.NewBuffer(shareData))
	if err != nil {
		return err
	}

	// 设置 Content-Type 和其他 header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	// 如果需要更多 headers，可以继续设置，比如 req.Header.Set("X-Custom-Header", "value")

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	// 关闭响应 body
	defer rsp.Body.Close()

	// 如果响应状态码 >= 400，则返回错误信息
	if rsp.StatusCode >= 400 {
		return fmt.Errorf("error sending shares: %v", string(data))
	}

	fmt.Println("shares upload success")
	return nil
}

func DecryptByUserID(userID int, authUrl string, token string) error {
	var identityName = Key.GetEntity().PrimaryIdentity().Name
	fmt.Printf("your identity_name is %v\n", identityName)

	req, err := http.NewRequest(
		"GET",
		authUrl+"/api/shamir/"+strconv.Itoa(userID)+"?identity_name="+url.QueryEscape(identityName),
		nil,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)

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
