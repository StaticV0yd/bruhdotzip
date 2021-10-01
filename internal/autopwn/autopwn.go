package autopwn

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/masterzen/winrm"
	"github.com/opt/red-script/internal/files"
	"golang.org/x/crypto/ssh"
)

// Will attempt to execute a script located at scriptPath on the target host using
//	provided user and password through SSH.
func SSHAutopwn(host, user, password, scriptPath string) {
	//defer wg.Done()
	// Set up SSH connection config
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		//HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
		Timeout: 10 * time.Second,
	}

	// Attempt SSH connection/login
	conn, err := ssh.Dial("tcp", host+":22", sshConfig)
	if err != nil {
		os.Stderr.WriteString("ERROR(autopwn): Could not connect to SSH on " + host + " with provided user and password.\n")
		os.Stderr.WriteString(err.Error() + "\n")
		return
	}
	fmt.Println("autopwn: Successful SSH connection @", host)
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		os.Stderr.WriteString("ERROR(autopwn): Could not create an SSH session.\n")
		os.Stderr.WriteString(err.Error() + "\n")
		return
	}
	fmt.Println("autopwn: Successful SSH session creation @", host)

	fileString, err := files.ReadString(scriptPath)
	fmt.Println(fileString)
	err = session.Run("echo \"" + fileString + "\"> /tmp/output.sh")
	if err != nil {
		os.Stderr.WriteString("ERROR(autopwn): Failed to write script to a file on the remote host.\n")
		os.Stderr.WriteString(err.Error() + "\n")
		return
	}
	session.Close()
	session, err = conn.NewSession()
	if err != nil {
		os.Stderr.WriteString("ERROR(autopwn): Could not create an SSH session.\n")
		os.Stderr.WriteString(err.Error() + "\n")
		session.Close()
		return
	}
	fmt.Println("autopwn: Successful SSH session creation @", host)
	defer session.Close()
	err = session.Run("sh /tmp/output.sh")
	if err != nil {
		os.Stderr.WriteString("ERROR(autopwn): Failed to execute script on the remote host.\n")
		os.Stderr.WriteString(err.Error() + "\n")
		return
	}

	// Close the connection when the rest of the function is done running

	// TODO: Read script from scriptPath
	// TODO: Execute script through SSH on host

}

// Will attempt to execute a script located at scriptPath on the target host using
//	provided user and password through WinRM.
func WinRMAutopwn(host, user, password, scriptPath string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Split host and port number to be used when creating an endpoint
	splitHost := strings.Split(host, ":")
	port, err := strconv.Atoi(splitHost[1])
	if err != nil {
		os.Stderr.WriteString("ERROR(autopwn): Failed to convert port number into int.\n")
		os.Stderr.WriteString(err.Error() + "\n")
		return
	}

	// Create an endpoint and setup the WinRM connection
	endpoint := winrm.NewEndpoint(splitHost[0], port, false, true, nil, nil, nil, 30*time.Second)
	params := winrm.DefaultParameters
	params.TransportDecorator = func() winrm.Transporter { return &winrm.ClientNTLM{} }

	// Attempt to create WinRM client
	client, err := winrm.NewClientWithParameters(endpoint, user, password, params)
	if err != nil {
		os.Stderr.WriteString("ERROR(autopwn): Failed to create WinRM client.\n")
		os.Stderr.WriteString(err.Error() + "\n")
		return
	}

	// Attempt to execute a basic powershell command through WinRM
	cmd := winrm.Powershell("ipconfig")
	_, err = client.Run(cmd, os.Stdout, os.Stderr)
	if err != nil {
		os.Stderr.WriteString("ERROR(autopwn): Failed to execute command through WinRM.\n")
		os.Stderr.WriteString(err.Error() + "\n")
		return
	}

	fmt.Println("Successful WinRM connection @", host)

	// TODO: Read script from scriptPath
	// TODO: Execute script through WinRM on host
}
