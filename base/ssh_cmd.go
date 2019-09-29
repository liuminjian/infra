package base

func RunCmd(h *Host, cmd string, envs ...EnvMap) (output string, err error) {
	conn, err := NewSSHClient(h)
	defer conn.Close()
	if err != nil {
		return
	}
	output, err = runCommand(conn, cmd, envs...)
	return
}

func RunSudoCmd(h *Host, cmd string, envs ...EnvMap) (output string, err error) {
	conn, err := NewSSHClient(h)
	defer conn.Close()
	if err != nil {
		return
	}
	output, err = sudoCommand(conn, cmd, h.User, h.Password, envs...)
	return
}
