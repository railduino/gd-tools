package agent

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
)

func (req *Request) Send() error {
	if req.Conn == nil {
		return fmt.Errorf("oops - no req.Conn?")
	}

	if req.Dry {
		fmt.Println("[dry]", req)
		return nil
	}

	if req.Verbose {
		fmt.Printf("[run] Request: '%v'\n", req)
	}

	encoder := json.NewEncoder(req.Conn)
	if err := encoder.Encode(req); err != nil {
		return err
	}

	var resp Response
	decoder := json.NewDecoder(req.Conn)
	if err := decoder.Decode(&resp); err != nil {
		return err
	}
	if resp.Err != "" {
		return fmt.Errorf("response returned with error: %s", resp.Err)
	}

	if len(resp.UserIDs) > 0 {
		if req.Verbose {
			fmt.Printf("[run] UserIDs: '%+v'\n", resp.UserIDs)
		}
		for _, userID := range resp.UserIDs {
			if err := SaveUserIDs(userID); err != nil {
				return err
			}
		}
	}

	if req.RustDesk != nil && resp.RustDesk != nil {

		reqPub := req.RustDesk.GetPublic()
		respPub := resp.RustDesk.GetPublic()
		if reqPub != "" && respPub != "" && reqPub != respPub {
			return fmt.Errorf("rustdesk public key mismatch")
		}

		if req.RustDesk.PrivateB64 != "" &&
			resp.RustDesk.PrivateB64 != "" &&
			req.RustDesk.PrivateB64 != resp.RustDesk.PrivateB64 {
			return fmt.Errorf("rustdesk private key mismatch")
		}

		if err := resp.RustDesk.Save(); err != nil {
			return err
		}
	}

	if req.Verbose {
		fmt.Printf("[run] Response: '%v'\n", &resp)
	} else if !req.Quiet {
		for _, line := range resp.Result {
			fmt.Println("[run]", line)
		}
	}

	return nil
}

func (req *Request) SendToAgent(conn *tls.Conn, debug, dry bool) error {
	if conn == nil {
		return fmt.Errorf("oops - no conn?")
	}

	if dry {
		fmt.Println("[dry]", req)
		return nil
	}

	if debug {
		fmt.Printf("[run] Request: '%v'\n", req)
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(req); err != nil {
		return err
	}

	var resp Response
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&resp); err != nil {
		return err
	}
	if resp.Err != "" {
		return fmt.Errorf("response returned with error: %s", resp.Err)
	}

	if len(resp.UserIDs) > 0 {
		if debug {
			fmt.Printf("[run] UserIDs: '%+v'\n", resp.UserIDs)
		}
		for _, userID := range resp.UserIDs {
			if err := SaveUserIDs(userID); err != nil {
				return err
			}
		}
	}

	if debug {
		fmt.Printf("[run] Response: '%v'\n", &resp)
	} else {
		for _, line := range resp.Result {
			fmt.Println("[run]", line)
		}
	}

	return nil
}
