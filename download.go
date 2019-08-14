package mget

func Download(url, path string, nWork int) error {
	task, err := NewTask(url, path, nWork)
	if err != nil {
		return err
	}

	done := task.drawDaemon()
	err = task.exec()
	<-done
	return err
}
