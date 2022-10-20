package ftpup

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"example.com/config"
	"example.com/dbhandler"
	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func FtpUploadDownload(ftpConf config.Ftp) {

	// Get user name and pass
	user := ftpConf.Auth.User
	pass := ftpConf.Auth.Password

	// Parse Host and Port
	host := ftpConf.Auth.Host
	// Default SFTP port
	port := 22

	//hostKey := getHostKey(host)

	fmt.Fprintf(os.Stdout, "Connecting to %s ...\n", host)

	var auths []ssh.AuthMethod

	// Try to use $SSH_AUTH_SOCK which contains the path of the unix file socket that the sshd agent uses
	// for communication with other processes.
	if aconn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(aconn).Signers))
	}

	// Use password authentication if provided
	if pass != "" {
		auths = append(auths, ssh.Password(pass))
	}

	// Initialize client configuration
	config := ssh.ClientConfig{
		User: user,
		Auth: auths,
		// Uncomment to ignore host key check
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		//HostKeyCallback: ssh.FixedHostKey(hostKey),
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	// Connect to server
	conn, err := ssh.Dial("tcp", addr, &config)
	log.Println("Connected")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connecto to [%s]: %v\n", addr, err)
		os.Exit(1)
	}

	defer conn.Close()

	// Create new SFTP client
	sc, err := sftp.NewClient(conn)
	log.Println("SFTP client started")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start SFTP subsystem: %v\n", err)
		os.Exit(1)
	}
	defer sc.Close()

	filesToUpload, err := getFilesToUpload(ftpConf.UploadFolder)
	if err != nil {
		log.Fatal("Application was not able to get list of files which should be uploaded")
		log.Fatal(err)
	}

	for _, file := range filesToUpload {
		localFile := filepath.Join(ftpConf.UploadFolder, file)
		remoteFile := filepath.Join(ftpConf.RemoteUp, file)
		localFileArchive := filepath.Join(ftpConf.UploadArchive, file)
		uploadFile(sc, localFile, remoteFile)
		moveFile(localFile, localFileArchive)
	}

	filesToDownload, err := getFilesToDownload(sc, ftpConf.RemoteDown)
	if err != nil {
		log.Fatal("Application was not able to get list of files which should be downloaded")
		log.Fatal(err)
	}

	for _, file := range filesToDownload {
		_, attachmentUuid := splitExportedFilename(file)
		if dbhandler.AttachmentGetStatus(attachmentUuid) == 0 {
			localFile := filepath.Join(ftpConf.DownloadFolder, file)
			remoteFile := filepath.Join(ftpConf.RemoteDown, file)
			downloadFile(sc, remoteFile, localFile, attachmentUuid)
		}
	}
}

func splitExportedFilename(filename string) (string, string) {
	stringArray := strings.Split(filename, "_")
	var dateOfExport string
	var attachmentUuid string
	if len(stringArray) == 1 {
		dateOfExport = ""
		attachmentUuid = strings.Split(stringArray[0], ".")[0]
	} else {
		dateOfExport = stringArray[0]
		attachmentUuid = strings.Split(stringArray[1], ".")[0]
	}

	return dateOfExport, attachmentUuid
}

// Get host key from local known hosts
func getHostKey(host string) ssh.PublicKey {
	// parse OpenSSH known_hosts file
	// ssh or use ssh-keyscan to get initial key
	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read known_hosts file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var hostKey ssh.PublicKey
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}
		if strings.Contains(fields[0], host) {
			var err error
			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing %q: %v\n", fields[2], err)
				os.Exit(1)
			}
			break
		}
	}

	if hostKey == nil {
		fmt.Fprintf(os.Stderr, "No hostkey found for %s", host)
		os.Exit(1)
	}

	return hostKey
}

func getFilesToUpload(dirName string) ([]string, error) {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var filesToUpload []string

	for _, file := range files {
		if !file.IsDir() {
			fmt.Println(file.Name())
			filesToUpload = append(filesToUpload, file.Name())
		}

	}

	return filesToUpload, err
}

func getFilesToDownload(sc *sftp.Client, remoteDir string) ([]string, error) {
	fmt.Fprintf(os.Stdout, "Listing [%s] ...\n\n", remoteDir)
	var filesToDownload []string

	files, err := sc.ReadDir(remoteDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to list remote dir: %v\n", err)
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() {
			filesToDownload = append(filesToDownload, file.Name())
		}
	}

	log.Println(filesToDownload)

	return filesToDownload, nil
}

func uploadFile(sc *sftp.Client, localFile, remoteFile string) (err error) {
	fmt.Fprintf(os.Stdout, "Uploading [%s] to [%s] ...\n", localFile, remoteFile)

	srcFile, err := os.Open(localFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open local file: %v\n", err)
		return
	}
	defer srcFile.Close()

	// // Make remote directories recursion
	// parent := filepath.Dir(remoteFile)
	// path := string(filepath.Separator)
	// dirs := strings.Split(parent, path)
	// for _, dir := range dirs {
	// 	path = filepath.Join(path, dir)
	// 	sc.Mkdir(path)
	// }

	// Note: SFTP To Go doesn't support O_RDWR mode
	dstFile, err := sc.OpenFile(remoteFile, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open remote file: %v\n", err)
		return
	}
	defer dstFile.Close()

	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to upload local file: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%d bytes copied\n", bytes)

	return
}

func archiveUploads(localFile string) {
	return
}

func moveFile(localFile, localFileArchive string) {
	err := os.Rename(localFile, localFileArchive)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("File %v was moved to %v", localFile, localFileArchive)
	return
}

func downloadFile(sc *sftp.Client, remoteFile, localFile, attachmentUuid string) (err error) {
	fmt.Fprintf(os.Stdout, "Downloading [%s] to [%s] ...\n", remoteFile, localFile)
	// Note: SFTP To Go doesn't support O_RDWR mode
	srcFile, err := sc.OpenFile(remoteFile, (os.O_RDONLY))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open remote file: %v\n", err)
		return
	}
	defer srcFile.Close()

	dstFile, err := os.Create(localFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open local file: %v\n", err)
		return
	}
	defer dstFile.Close()

	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to download remote file: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%d bytes copied\n", bytes)
	dbhandler.AttachmentIsProcessed(attachmentUuid)
	dbhandler.AllAttachmentsComplete(attachmentUuid)

	return
}
