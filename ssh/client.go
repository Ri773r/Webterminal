package ssh

import (
	"fmt"
	"io/ioutil"
	"net"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

// Client Ssh Client wrapper
type Client struct {
	*gossh.Client
	Cols int
	Rows int
}

// NewSSHClient Return Client wrapper object
func NewSSHClient(user, password, host, key string, port int, cipherList []string) (*Client, error) {
	var (
		auth         []gossh.AuthMethod
		addr         string
		config       gossh.Config
		clientConfig *gossh.ClientConfig
		client       *gossh.Client
		err          error
	)

	if key == "" {
		// password authentication
		auth = append(auth, gossh.Password(password))
	} else {
		// key authentication
		pemBytes, err := ioutil.ReadFile(key)
		if err != nil {
			return nil, err
		}
		var signer gossh.Signer
		if password == "" {
			// parse key with password
			signer, err = gossh.ParsePrivateKey(pemBytes)
			if err != nil {
				return nil, err
			}
		} else {
			// parse key without password
			signer, err = gossh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(password))
			if err != nil {
				return nil, err
			}
		}
		auth = append(auth, gossh.PublicKeys(signer))
	}

	// allowed cipher algorithms
	if len(cipherList) == 0 {
		config = gossh.Config{
			Ciphers: []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"},
		}
	} else {
		config = gossh.Config{
			Ciphers: cipherList,
		}
	}

	// authentication callback function
	hostKeyCallback := func(hostname string, remote net.Addr, key gossh.PublicKey) error {
		return nil
	}
	clientConfig = &gossh.ClientConfig{
		User:            user,
		Auth:            auth,
		Config:          config,
		Timeout:         30 * time.Second,
		HostKeyCallback: hostKeyCallback,
	}
	addr = fmt.Sprintf("%s:%d", host, port)
	if client, err = gossh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	return &Client{Client: client}, nil
}

// GetDefaultTerminalOption get terminal default mode
func GetDefaultTerminalOption() gossh.TerminalModes {
	return gossh.TerminalModes{
		gossh.ECHO:          1,
		gossh.TTY_OP_ISPEED: 14400,
		gossh.TTY_OP_OSPEED: 14400,
	}
}
