package pkg

type Logger []string

func Log(logger Logger, message string) {
	logger = append(logger, message)
}
