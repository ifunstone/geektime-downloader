package progress

type Download struct {
	CourseTitle     string
	ItemTitle       string
	Stage           string
	CurrentItem     int
	TotalItems      int
	DownloadedBytes int64
	TotalBytes      int64
}

type Func func(Download)
