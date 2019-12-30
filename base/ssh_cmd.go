package base

import "golang.org/x/crypto/ssh"

func RunCmd(h *Host, cfg ssh.Config, cmd string, envs ...EnvMap) (output string, err error) {
	conn, err := NewSSHClient(h, cfg)
	defer conn.Close()
	if err != nil {
		return
	}
	output, err = runCommand(conn, cmd, envs...)
	return
}

func RunSudoCmd(h *Host, cfg ssh.Config, cmd string, envs ...EnvMap) (output string, err error) {
	conn, err := NewSSHClient(h, cfg)
	defer conn.Close()
	if err != nil {
		return
	}
	output, err = sudoCommand(conn, cmd, h.User, h.Password, envs...)
	return
}

//判断文件是否存在
func FileExist(h *Host, cfg ssh.Config, path string) error {
	client, err := NewSftpClient(h, cfg)
	defer client.Close()
	if err != nil {
		return err
	}
	_, err = client.Lstat(path)
	return err
}
