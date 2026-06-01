package uiapp

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
)

const maxUILogBytes int64 = 128 * 1024

func readLogFile() string {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "无法读取日志目录"
	}
	logFilePath := filepath.Join(userConfigDir, logger.GeektimeLogFolder, logger.GeektimeLogFolder+".log")
	data, truncated, err := readTailFile(logFilePath, maxUILogBytes)
	if err != nil {
		if os.IsNotExist(err) {
			return "当前还没有日志文件"
		}
		return "读取日志失败: " + err.Error()
	}
	if len(data) == 0 {
		return "日志文件为空"
	}
	if truncated {
		return "[日志过大，仅显示最后 128KB]\n" + string(data)
	}
	return string(data)
}

func readTailFile(path string, maxBytes int64) ([]byte, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, false, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, false, err
	}

	size := info.Size()
	if size == 0 {
		return []byte{}, false, nil
	}

	offset := int64(0)
	truncated := false
	if size > maxBytes {
		offset = size - maxBytes
		truncated = true
	}

	buf := make([]byte, size-offset)
	if _, err := file.ReadAt(buf, offset); err != nil {
		return nil, false, err
	}

	if truncated {
		if idx := strings.IndexByte(string(buf), '\n'); idx >= 0 && idx+1 < len(buf) {
			buf = buf[idx+1:]
		}
	}

	return buf, truncated, nil
}
