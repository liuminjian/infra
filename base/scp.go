package base

import (
	"github.com/pkg/sftp"
	"github.com/prometheus/common/log"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const maxPacket = 1 << 15

func NewSftpClient(h *Host, cfg ssh.Config) (*sftp.Client, error) {
	conn, err := NewSSHClient(h, cfg)
	if err != nil {
		return nil, err
	}
	return sftp.NewClient(conn, sftp.MaxPacket(maxPacket))
}

func ScpPut(h *Host, cfg ssh.Config, localPath, remotePath string) error {
	client, err := NewSftpClient(h, cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	return putFile(client, localPath, remotePath)
}

func putFile(client *sftp.Client, localPath, remotePath string) error {
	info, err := os.Lstat(localPath)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return putLinkFile(client, localPath, remotePath)
	}
	if info.IsDir() {
		return putDirectory(client, localPath, remotePath)
	}
	return putLocalFile(client, localPath, remotePath, info)
}

func putLocalFile(client *sftp.Client, localPath, remotePath string, info os.FileInfo) error {
	localFile, err := os.Open(localPath)
	if err != nil {
		log.Error(err)
		return err
	}
	defer localFile.Close()
	err = client.MkdirAll(filepath.Dir(remotePath))
	if err != nil {
		log.Error(err)
		return err
	}
	remoteFile, err := client.Create(remotePath)
	if err != nil {
		log.Error(err)
		return err
	}
	defer remoteFile.Close()
	err = client.Chmod(remoteFile.Name(), info.Mode())
	if err != nil {
		log.Error(err)
		return err
	}
	size, err := io.Copy(remoteFile, localFile)
	log.Debugf("put file %s -> %s %d", localPath, remotePath, size)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func putLinkFile(client *sftp.Client, localPath, remotePath string) error {
	readLocal, err := os.Readlink(localPath)
	if err != nil {
		return err
	}
	return putFile(client, readLocal, remotePath)
}

func putDirectory(client *sftp.Client, localPath, remotePath string) error {
	contents, err := ioutil.ReadDir(localPath)
	if err != nil {
		log.Error(err)
		return err
	}
	for _, content := range contents {
		src := filepath.Join(localPath, content.Name())
		dst := filepath.Join(remotePath, content.Name())
		err := putFile(client, src, dst)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

func ScpGet(h *Host, cfg ssh.Config, localPath, remotePath string) error {
	client, err := NewSftpClient(h, cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	return getFile(client, localPath, remotePath)
}

func getFile(client *sftp.Client, localPath, remotePath string) error {
	info, err := client.Lstat(remotePath)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return getLinkFile(client, localPath, remotePath)
	}
	if info.IsDir() {
		return getDirectory(client, localPath, remotePath)
	}
	return getRemoteFile(client, localPath, remotePath, info)
}

func getRemoteFile(client *sftp.Client, localPath, remotePath string, info os.FileInfo) error {
	remoteFile, err := client.Open(remotePath)
	if err != nil {
		log.Error(err)
		return err
	}
	defer remoteFile.Close()
	localFile, err := os.Create(localPath)
	if err != nil {
		log.Error(err)
		return err
	}
	defer localFile.Close()

	size, err := io.Copy(localFile, remoteFile)
	log.Debugf("get file %s -> %s %d", remotePath, localPath, size)
	if err != nil {
		log.Error(err)
		return err
	}

	err = os.Chmod(localPath, info.Mode())
	return err
}

func getLinkFile(client *sftp.Client, localPath, remotePath string) error {
	readRemote, err := client.ReadLink(remotePath)
	if err != nil {
		return err
	}
	return getFile(client, localPath, readRemote)
}

func getDirectory(client *sftp.Client, localPath, remotePath string) error {
	contents, err := client.ReadDir(remotePath)
	if err != nil {
		log.Error(err)
		return err
	}
	for _, content := range contents {
		src := filepath.Join(remotePath, content.Name())
		dst := filepath.Join(localPath, content.Name())
		err := getFile(client, dst, src)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}
