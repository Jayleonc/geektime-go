package logger

func Error(err error) Field {
	return Field{"error", err}
}

func Int64(key string, val int64) Field {
	return Field{key, val}
}

func String(key, val string) Field {
	return Field{key, val}
}

func Bool(key string, val bool) Field {
	return Field{key, val}
}
